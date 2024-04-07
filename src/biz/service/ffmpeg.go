package service

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/helper"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	custff "github.com/CE-Thesis-2023/ltd/src/internal/ffmpeg"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/CE-Thesis-2023/ltd/src/models/ms"
	"github.com/CE-Thesis-2023/ltd/src/models/rest"
	"go.uber.org/zap"
)

type onGoingProcess struct {
	SourceUrl      string
	DestinationUrl string
	proc           *exec.Cmd
}

func (c *onGoingProcess) Cancel(ctx context.Context) error {
	if c.proc != nil {
		return c.proc.Cancel()
	}
	return nil
}

type mediaService struct {
	onGoingProcesses map[string]*onGoingProcess
	mu               sync.Mutex
}

func NewMediaService() MediaServiceInterface {
	return &mediaService{
		onGoingProcesses: map[string]*onGoingProcess{},
	}
}

func (s *mediaService) Shutdown() {
	logger.SInfo("shutting down media service")
	for cameraId, p := range s.onGoingProcesses {
		if p.proc != nil {
			if p.proc.Cancel != nil {
				if err := p.proc.Cancel(); err != nil {
					logger.SDebug("error shutting down FFmpeg process", zap.Error(err))
					continue
				}
				delete(s.onGoingProcesses, cameraId)
				logger.SDebug("shutdown FFmpeg process", zap.String("cameraId", cameraId))
			}
		}
	}
	logger.SDebug("released streaming pool")
}

type MediaServiceInterface interface {
	RequestFFmpegRtspToSrt(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error
	CancelFFmpegRtspToSrt(ctx context.Context, camera *db.Camera) error
	ListOngoingStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error)
	Shutdown()
}

func (s *mediaService) RequestFFmpegRtspToSrt(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	logger.SInfo("requested starting to perform RTSP to SRT transcoding stream", zap.String("request", req.CameraId))

	sourceUrl := s.buildRtspStreamUrl(camera)
	logger.SDebug("RTSP source stream URL",
		zap.String("source", sourceUrl))

	destinationUrl := s.buildPushSrtUrl(&ms.PushStreamingRequest{
		StreamName: req.CameraId,
	})
	logger.SDebug("SRT destination stream URL",
		zap.String("destination", destinationUrl))

	if s.isThisStreamGoing(camera) {
		logger.SInfo("RequestFFmpegRtspToSrt: stream already exists")
		return nil
	}

	command := s.buildFfmpegRestreamingCommand(sourceUrl, destinationUrl)
	logger.SDebug("transcoding stream FFmpeg command",
		zap.String("command", command.String()))
	if command == nil {
		logger.SError("failed to build FFmpeg command")
		return custerror.FormatInternalError("failed to build FFmpeg os/exec command")
	}

	if err := helper.Do(func() error {
		s.recordThisStream(camera,
			sourceUrl,
			destinationUrl,
			command)

		logger.SInfo("starting transcoding stream")
		if err := command.Start(); err != nil {
			logger.SError("failed to run FFmpeg process", zap.Error(err))
			return err
		}
		logger.SInfo("transcoding stream ended")

		return nil
	},
		helper.Attempts(3),
		helper.RetryIf(func(err error) bool {
			if s.shouldRestartStream(err, camera) {
				logger.SInfo("will restart transcoding stream",
					zap.String("camera_id", camera.CameraId))
				return true
			}
			logger.SInfo("transcoding stream will not restart",
				zap.String("camera_id", camera.CameraId))
			return false
		})); err != nil {
		logger.SError("failed to start transcoding stream",
			zap.Error(err))
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.onGoingProcesses, req.CameraId)
	}
	return nil
}

func (s *mediaService) buildFfmpegRestreamingCommand(sourceUrl string, destinationUrl string) *exec.Cmd {
	configs := configs.Get()
	var binPath string
	var err error
	if configs.Ffmpeg.BinaryPath != "" {
		binPath, err = filepath.Abs(configs.Ffmpeg.BinaryPath)
		if err != nil {
			logger.SFatal("missing FFmpeg binary path", zap.Error(err))
			return nil
		}
	}

	cmd := custff.NewFFmpegCommand()
	cmd.WithSourceUrl(sourceUrl).
		WithBinPath(binPath).
		WithGlobalArguments(
			map[string]string{
				"hide_banner": "",
				"loglevel":    "info",
				"threads":     "2",
			},
		).
		WithInputArguments(map[string]string{
			"avoid_negative_ts":           "make_zero",
			"fflags":                      "+genpts+discardcorrupt",
			"rtsp_transport":              "tcp",
			"use_wallclock_as_timestamps": "1",
			"timeout":                     "5000000",
		}).
		WithDestinationUrl(destinationUrl).
		WithOutputArguments(map[string]string{
			"f":        "mpegts",
			"c:v":      "libx264",
			"preset:v": "faster",
			"tune:v":   "zerolatency",
		}).
		WithScale(20, 1280, 720).
		WithHardwareAccelerationType(custff.VA_API)

	execCmd, err := cmd.String()
	if err != nil {
		logger.SError("failed to build FFmpeg command", zap.Error(err))
		return nil
	}

	logger.SDebug("FFmpeg command", zap.String("command", execCmd))

	return exec.Command("/bin/bash", "-c", execCmd)
}

func (s *mediaService) shouldRestartStream(err error, camera *db.Camera) bool {
	_, found := s.onGoingProcesses[camera.CameraId]
	if found && err != nil {
		return true
	}
	return false
}

func (s *mediaService) recordThisStream(camera *db.Camera, sourceUrl string, destinationUrl string, proc *exec.Cmd) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onGoingProcesses[camera.CameraId] = &onGoingProcess{
		SourceUrl:      sourceUrl,
		DestinationUrl: destinationUrl,
		proc:           proc,
	}
}

func (s *mediaService) isThisStreamGoing(camera *db.Camera) bool {
	pr, found := s.onGoingProcesses[camera.CameraId]
	if found {
		logger.SDebug("transcoding stream already going",
			zap.Reflect("process", pr))
		return true
	}
	return false
}

func (s *mediaService) CancelFFmpegRtspToSrt(ctx context.Context, camera *db.Camera) error {
	logger.SInfo("requested canceling RTSP to SRT transcoding stream",
		zap.String("camera_id", camera.CameraId))

	onGoingProcess, yes := s.onGoingProcesses[camera.CameraId]
	if !yes {
		logger.SError("transcoding stream not found",
			zap.String("camera_id", camera.CameraId))
		return nil
	}
	delete(s.onGoingProcesses, camera.CameraId)

	logger.SDebug("canceling transcoding stream",
		zap.String("camera_id", camera.CameraId))
	if err := onGoingProcess.Cancel(ctx); err != nil {
		logger.SError("failed to cancel transcoding stream",
			zap.String("camera_id", camera.CameraId),
			zap.Error(err))
		return err
	}

	logger.SInfo("canceled transcoding stream",
		zap.String("camera_id", camera.CameraId))
	return nil
}

func (s *mediaService) buildRtspStreamUrl(camera *db.Camera) string {
	u := &url.URL{}
	u.Scheme = "rtsp"
	u.Host = camera.Ip
	if camera.Port != 0 {
		u.Host = fmt.Sprintf("%s:%d", camera.Ip, camera.Port)
	}
	u = u.JoinPath("/ISAPI", "/Streaming", "channels", "101")
	u.User = url.UserPassword(camera.Username, camera.Password)
	url := u.String()
	logger.SDebug("buildRtspStreamUrl: stream url", zap.String("url", url))
	return url
}

func (s *mediaService) ListOngoingStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error) {
	logger.SDebug("ListOngoingStreams: request received")
	streams := s.onGoingProcesses
	resp := &rest.DebugListStreamsResponse{}
	for cameraId, s := range streams {
		resp.Streams = append(resp.Streams, rest.StreamInfo{
			CameraId:       cameraId,
			SourceUrl:      s.SourceUrl,
			DestinationUrl: s.DestinationUrl,
		})
	}
	logger.SDebug("ListOngoingStreams: streams", zap.Reflect("streams", resp))
	return resp, nil
}
