package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
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

type OpenGateProcess struct {
	proc           *exec.Cmd
	configs        *configs.OpenGateConfigs
	absComposePath string
	settings       []byte
}

type MediaServiceInterface interface {
	StartTranscodingStream(ctx context.Context, p *Process) error
	EndTranscodingStream(ctx context.Context, p *Process) error
	ComposeRestartOpenGate(ctx context.Context, p *OpenGateProcess) error
	ComposeUpOpenGate(ctx context.Context, p *OpenGateProcess) error
	ComposeDownOpenGate(ctx context.Context, p *OpenGateProcess) error
	DockerPullImages(ctx context.Context, images []string) error
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

	command := s.buildFfmpegRestreamingCommand(ctx, sourceUrl, destinationUrl)
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

func (s *mediaService) buildFfmpegRestreamingCommand(ctx context.Context, sourceUrl string, destinationUrl string) *exec.Cmd {
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

	return exec.CommandContext(ctx,
		"/bin/bash", "-c", execCmd)
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

func (s *mediaService) ComposeUpOpenGate(ctx context.Context, p *OpenGateProcess) error {
	logger.SInfo("requested to start OpenGate")

	if err := s.writeConfigurationFile(p); err != nil {
		logger.SError("failed to write configuration file", zap.Error(err))
		return err
	}

	command := s.buildOpenGateCommand(ctx, p)
	logger.SDebug("opening gate command",
		zap.String("command", command.String()))
	if command == nil {
		logger.SError("failed to build os/exec command")
		return custerror.FormatInternalError("failed to build os/exec command")
	}
	p.proc = command

	logger.SInfo("starting OpenGate")
	if err := command.Run(); err != nil {
		logger.SError("failed to run process", zap.Error(err))
		return err
	}
	logger.SInfo("OpenGate started")
	return nil
}

func (s *mediaService) buildOpenGateCommand(ctx context.Context, p *OpenGateProcess) *exec.Cmd {
	filePath := p.configs.DockerComposePath
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		logger.SError("failed to get absolute path", zap.Error(err))
		return nil
	}
	p.absComposePath = absPath

	cmd := exec.CommandContext(ctx,
		"docker", "compose", "-f", absPath, "up", "-d")
	return cmd
}

func (s *mediaService) writeConfigurationFile(p *OpenGateProcess) error {
	filePath := p.configs.ConfigurationPath
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		logger.SError("failed to get absolute path", zap.Error(err))
		return err
	}

	f, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.SError("failed to open file", zap.Error(err))
		return err
	}
	defer f.Close()

	data := p.settings
	if _, err := f.Write(data); err != nil {
		logger.SError("failed to write to file", zap.Error(err))
		return err
	}

	return nil
}

func (s *mediaService) ComposeDownOpenGate(ctx context.Context, p *OpenGateProcess) error {
	if p.absComposePath == "" {
		filePath := p.configs.DockerComposePath
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			logger.SError("failed to get absolute path", zap.Error(err))
			return nil
		}
		p.absComposePath = absPath
	}
	if p.proc != nil {
		if p.proc.Process != nil {
			if err := p.proc.Process.Signal(syscall.SIGTERM); err != nil {
				logger.SDebug("no other Compose process running")
			}
			logger.SDebug("interrupted Compose up process")
		}
	}

	cmd := s.buildComposeDownCommand(ctx, p)
	if err := cmd.Run(); err != nil {
		logger.SError("failed to run process", zap.Error(err))
		return nil
	}

	logger.SDebug("Compose down process completed")
	return nil
}

func (s *mediaService) buildComposeDownCommand(_ context.Context, p *OpenGateProcess) *exec.Cmd {
	cmd := exec.Command(
		"docker", "compose", "-f", p.absComposePath, "down")
	return cmd
}

func (s *mediaService) ComposeRestartOpenGate(ctx context.Context, p *OpenGateProcess) error {
	if p.absComposePath == "" {
		filePath := p.configs.DockerComposePath
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			logger.SError("failed to get absolute path", zap.Error(err))
			return err
		}
		p.absComposePath = absPath
	}

	cmd := exec.CommandContext(ctx,
		"docker", "compose", "-f", p.absComposePath, "restart")
	if err := cmd.Run(); err != nil {
		logger.SError("failed to run process", zap.Error(err))
		return err
	}

	logger.SDebug("Compose restart process completed")
	return nil
}

func (s *mediaService) DockerPullImages(ctx context.Context, images []string) error {
	var wg sync.WaitGroup
	wg.Add(len(images))
	for _, i := range images {
		go func(image string) {
			cmd := s.buildDockerPullCmd(ctx, image)
			if err := cmd.Run(); err != nil {
				logger.SError("failed to run process",
					zap.Error(err))
			}
			defer wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

func (s *mediaService) buildDockerPullCmd(ctx context.Context, image string) *exec.Cmd {
	return exec.CommandContext(ctx,
		"docker", "pull", image)
}
