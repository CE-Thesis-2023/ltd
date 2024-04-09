package service

import (
	"context"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	custff "github.com/CE-Thesis-2023/ltd/src/internal/ffmpeg"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"
)

type mediaService struct {
}

func NewMediaService() MediaServiceInterface {
	return &mediaService{}
}

type Process struct {
	proc     *exec.Cmd
	cameraId string
	configs  *web.TranscoderStreamConfiguration
}

type MediaServiceInterface interface {
	StartTranscodingStream(ctx context.Context, p *Process) error
	EndTranscodingStream(ctx context.Context, p *Process) error
}

func (s *mediaService) StartTranscodingStream(ctx context.Context, p *Process) error {
	logger.SInfo("requested starting to perform RTSP to SRT transcoding stream",
		zap.String("request", p.cameraId))

	sourceUrl := p.configs.SourceUrl
	logger.SDebug("RTSP source stream URL",
		zap.String("source", sourceUrl))

	destinationUrl := p.configs.PublishUrl
	logger.SDebug("SRT destination stream URL",
		zap.String("destination", destinationUrl))

	command := s.buildFfmpegRestreamingCommand(sourceUrl, destinationUrl)
	logger.SDebug("transcoding stream FFmpeg command",
		zap.String("command", command.String()))
	if command == nil {
		logger.SError("failed to build FFmpeg command")
		return custerror.FormatInternalError("failed to build FFmpeg os/exec command")
	}
	p.proc = command

	logger.SInfo("starting transcoding stream")
	if err := command.Run(); err != nil {
		logger.SError("failed to run FFmpeg process", zap.Error(err))
		return err
	}
	logger.SInfo("transcoding stream ended")
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

func (s *mediaService) EndTranscodingStream(ctx context.Context, p *Process) error {
	logger.SInfo("requested to end transcoding stream",
		zap.String("camera_id", p.cameraId))

	if p.proc == nil {
		logger.SInfo("no FFmpeg process to kill",
			zap.String("camera_id", p.cameraId))
		return nil
	}

	if p.proc.ProcessState != nil && p.proc.ProcessState.Exited() {
		logger.SInfo("FFmpeg process already exited",
			zap.String("camera_id", p.cameraId))
		return nil
	}

	if p.proc.Process == nil {
		logger.SInfo("FFmpeg process is nil",
			zap.String("camera_id", p.cameraId))
		return nil
	}

	if err := p.proc.Process.Signal(syscall.SIGTERM); err != nil {
		logger.SError("failed to kill FFmpeg process",
			zap.String("camera_id", p.cameraId),
			zap.Error(err))
		return err
	}

	logger.SInfo("FFmpeg process killed",
		zap.String("camera_id", p.cameraId))
	return nil
}
