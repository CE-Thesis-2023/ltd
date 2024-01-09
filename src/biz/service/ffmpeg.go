package service

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	custrtsp "github.com/CE-Thesis-2023/ltd/internal/rtsp"
	"github.com/CE-Thesis-2023/ltd/models/db"
	"github.com/CE-Thesis-2023/ltd/models/events"
	"os/exec"
	"path/filepath"

	"github.com/avast/retry-go"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"
)

func (s *mediaService) RequestFFmpegRtspToSrt(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	logger.SDebug("RequestFFmpegRtspToSrt", zap.String("request", req.CameraId))

	sourceUrl := s.buildRtspStreamUrl(camera, req)
	logger.SDebug("RequestFFmpegRtspToSrt: source RTSP",
		zap.String("source/", sourceUrl))

	destinationUrl := s.buildRtspStreamUrl(camera, req)
	logger.SDebug("RequestFFmpegRtspToSrt: destination SRT", zap.String("destination", destinationUrl))

	if s.isThisStreamGoing(ctx, camera, sourceUrl, destinationUrl) {
		logger.SInfo("RequestFFmpegRtspToSrt: stream already exists")
		return nil
	}

	command := s.buildFfmpegRestreamingCommand(sourceUrl, destinationUrl)
	logger.SDebug("RequestFFmpegRtspToSrt: commanđ", zap.String("command", command.String()))

	srcUrlParsed, err := base.ParseURL(sourceUrl)
	if err != nil {
		logger.SError("RequestFFmpegRtspToSrt: base.ParseURL of sourceUrl", zap.Error(err))
		return err
	}

	client := custrtsp.New()
	if err := client.Start(srcUrlParsed.Scheme, srcUrlParsed.Host); err != nil {
		logger.SError("RequestFFmpegRtspToSrt: client.Start", zap.Error(err))
		return err
	}
	defer client.Close()

	sess, _, err := client.Describe(srcUrlParsed)
	if err != nil {
		logger.SError("RequestFFmpegRtspToSrt: rtspClient.Describe", zap.Error(err))
		return err
	}

	logger.SInfo("RequestFFmpegRtspToSrt: source DESCRIBE information", logger.Json("describeResp", sess))
	s.streamingPool.Submit(func() {
		retry.Do(func() error {
			compiledGoCommand := command.Compile()

			s.recordThisStream(ctx, camera, sourceUrl, destinationUrl, compiledGoCommand)
			logger.SDebug("RequestFFmpegRtspToSrt: reported this stream into memory")

			logger.SDebug("RequestFFmpegRtspToSrt: start FFMPEG process")
			if err := compiledGoCommand.Run(); err != nil {
				logger.SError("RequestFFmpegRtspToSrt: FFMPEG process error", zap.Error(err))
				return err
			}
			logger.SInfo("RequestFFmpegRtspToSrt: FFMPEG process finished")

			return nil
		}, retry.Attempts(3),
			retry.RetryIf(func(err error) bool {
				if s.shouldRestartStream(err, camera, sourceUrl, destinationUrl) {
					logger.SInfo("RequestFFmpegRtspToSrt: restarting stream")
					return true
				}
				logger.SInfo("RequestFFmpegRtspToSrt: will not restart stream")
				return false
			}))
		logger.SInfo("RequestFFmpegRtspToSrt: stream attempted 3 times, will disappear")
		delete(s.onGoingProcesses, req.CameraId)
	})

	logger.SDebug("RequestFFmpegRtspToSrt: assigned task")
	return nil
}

func (s *mediaService) buildFfmpegRestreamingCommand(sourceUrl string, destinationUrl string) *ffmpeg_go.Stream {
	cmd := ffmpeg_go.Input(sourceUrl).
		Output(destinationUrl, ffmpeg_go.KwArgs{
			"c:v":            "libx264",
			"c:a":            "aac",
			"f":              "mpegts",
			"preset":         "veryfast",
			"tune":           "zerolatency",
			"profile:v":      "baseline",
			"s":              "1280x720",
			"filter:v":       "fps=24",
			"timeout":        5000000,
			"v":              "debug",
			"rtsp_transport": "tcp",
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

func (s *mediaService) shouldRestartStream(err error, camera *db.Camera, sourceUrl string, destinationUrl string) bool {
	logger.SDebug("shouldRestartStream",
		zap.String("source", sourceUrl),
		zap.String("destination", destinationUrl))
	_, found := s.onGoingProcesses[camera.Id]
	if found && err != nil {
		return true
	}
	return false
}

func (s *mediaService) recordThisStream(ctx context.Context, camera *db.Camera, sourceUrl string, destinationUrl string, proc *exec.Cmd) {
	s.onGoingProcesses[camera.Id] = &onGoingProcess{
		SourceUrl:      sourceUrl,
		DestinationUrl: destinationUrl,
		proc:           proc,
	}
}

func (s *mediaService) isThisStreamGoing(ctx context.Context, camera *db.Camera, sourceUrl string, destinationUrl string) bool {
	pr, found := s.onGoingProcesses[camera.Id]
	if pr != nil {
		logger.SDebug("isThisStreamGoing: stream already ongoing", zap.Any("process", pr))
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
	logger.SDebug("CancelFFmpegRtspToSrt: cancel", zap.String("cameraId", camera.Id))
	onGoingProcess, yes := s.onGoingProcesses[camera.Id]
	if !yes {
		logger.SDebug("CancelFFmpegRtspToSrt: stream already canceled or not found")
		return custerror.ErrorPermissionDenied
	}

	delete(s.onGoingProcesses, camera.Id)

	logger.SDebug("CancelFFmpegRtspToSrt: canceling stream process")
	if err := onGoingProcess.Cancel(ctx); err != nil {
		logger.SError("CancelFFmpegRtspToSrt: Cancel", zap.Error(err))
		return err
	}

	logger.SDebug("CancelFFmpegRtspToSrt: stream canceled", zap.String("cameraId", camera.Id))
	return nil
}
