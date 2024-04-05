package service

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"

	"github.com/CE-Thesis-2023/ltd/src/helper"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/CE-Thesis-2023/ltd/src/models/ms"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"
)

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

	go func() {
		helper.Do(func() error {
			compiledGoCommand := command.Compile()

			s.recordThisStream(camera,
				sourceUrl,
				destinationUrl,
				compiledGoCommand)

			logger.SInfo("starting transcoding stream")
			if err := compiledGoCommand.Run(); err != nil {
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
			}))
		delete(s.onGoingProcesses, req.CameraId)
	}()

	return nil
}

func (s *mediaService) buildFfmpegRestreamingCommand(sourceUrl string, destinationUrl string) *ffmpeg_go.Stream {
	cmd := ffmpeg_go.Input(sourceUrl, ffmpeg_go.KwArgs{
		"rtsp_transport": "tcp",
	}).
		Output(destinationUrl, ffmpeg_go.KwArgs{
			"c:v":      "libx264",
			"c:a":      "aac",
			"f":        "mpegts",
			"tune":     "zerolatency",
			"preset":   "faster",
			"s":        "1280x720",
			"filter:v": "fps=25",
			"timeout":  5000000,
		}).ErrorToStdOut().
		WithCpuCoreLimit(2)
	configs := configs.Get()
	if configs.Ffmpeg.BinaryPath != "" {
		absPath, err := filepath.Abs(configs.Ffmpeg.BinaryPath)
		if err != nil {
			logger.SError("buildFfmpegRestreamingCommand: filepath.Abs", zap.Error(err))
		} else {
			cmd.SetFfmpegPath(absPath)
		}
	}
	return cmd
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
	if pr != nil {
		logger.SDebug("isThisStreamGoing: stream already ongoing", zap.Reflect("process", pr))
		return true
	}
	if found {
		if pr.proc != nil {
			if pr.proc.ProcessState != nil {
				if pr.proc.ProcessState.Exited() || pr.proc.ProcessState.ExitCode() != 0 {
					logger.SDebug("isThisStreamGoing: process associated with it has already exited")
					return false
				} else {
					logger.SDebug("isThisStreamGoing: process are not terminated yet")
					return true
				}
			}
		}
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
