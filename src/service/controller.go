package service

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"
)

type MediaController struct {
	mu              sync.Mutex
	ffmpegStreams   map[string]web.TranscoderStreamConfiguration
	needConcilation []string
	needRemoval     []string
	running         map[string]*Process
	mediaService    MediaServiceInterface
}

func NewMediaController(mediaService MediaServiceInterface) *MediaController {
	return &MediaController{
		ffmpegStreams:   make(map[string]web.TranscoderStreamConfiguration),
		running:         make(map[string]*Process),
		needConcilation: make([]string, 0),
		needRemoval:     make([]string, 0),
		mediaService:    mediaService,
	}
}

func (c *MediaController) Reconcile(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logger.SDebug("media controller context cancelled")
			c.cleanup()
			logger.SDebug("media controller cleaned up")
			return nil
		default:
			updated := false

			c.mu.Lock()
			for _, cameraId := range c.needConcilation {
				updated = true
				if err := c.startOrRestart(cameraId); err != nil {
					logger.SError("error reconciling stream",
						zap.String("cameraId", cameraId),
						zap.Error(err))
					c.mu.Unlock()
					return err
				}
				logger.SDebug("reconciled stream",
					zap.String("cameraId", cameraId))
			}
			c.needConcilation = []string{}

			for _, cameraId := range c.needRemoval {
				updated = true
				if err := c.stop(cameraId); err != nil {
					logger.SDebug("error stopping stream",
						zap.String("cameraId", cameraId))
					c.mu.Unlock()
					return err
				}
				logger.SDebug("stopped stream",
					zap.String("cameraId", cameraId))
			}
			c.needRemoval = []string{}
			c.mu.Unlock()

			if updated {
				logger.SDebug("media controller reconciled")
			}

			time.Sleep(200 * time.Millisecond)
			continue
		}
	}
}
func (c *MediaController) cleanup() {
	for cameraId := range c.running {
		if err := c.stop(cameraId); err != nil {
			logger.SDebug("error stopping stream",
				zap.String("cameraId", cameraId))
		}
	}
}

func (c *MediaController) Register(s web.TranscoderStreamConfiguration) (updated bool, err error) {
	return c.register(s)
}

func (c *MediaController) Exists(cameraId string) bool {
	_, found := c.ffmpegStreams[cameraId]
	return found
}

func (c *MediaController) Deregister(cameraId string) {
	c.deregister(cameraId)
}

func (c *MediaController) deregister(cameraId string) {
	_, found := c.ffmpegStreams[cameraId]
	if !found {
		logger.SDebug("stream not found",
			zap.String("cameraId", cameraId))
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.markForRemoval(cameraId)
	delete(c.ffmpegStreams, cameraId)
	logger.SDebug("deregistered stream",
		zap.String("cameraId", cameraId))
}

func (c *MediaController) markForRemoval(cameraId string) {
	c.needRemoval = append(c.needRemoval, cameraId)
}

func (c *MediaController) register(s web.TranscoderStreamConfiguration) (updated bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	curr, found := c.ffmpegStreams[s.CameraId]
	if found {
		if !c.needReconcile(curr, s) {
			logger.SDebug("skipped reconciling stream",
				zap.String("cameraId", s.CameraId))
			return false, nil
		}
		c.ffmpegStreams[s.CameraId] = s
		c.markForReconcile(s.CameraId)
		logger.SDebug("updated stream",
			zap.String("cameraId", s.CameraId))
		return true, nil
	}
	c.ffmpegStreams[s.CameraId] = s
	c.markForReconcile(s.CameraId)
	logger.SDebug("registered new stream",
		zap.String("cameraId", s.CameraId))
	return true, nil
}

func (c *MediaController) needReconcile(old web.TranscoderStreamConfiguration, new web.TranscoderStreamConfiguration) bool {
	if old.SourceUrl != new.SourceUrl {
		return true
	}
	if old.PublishUrl != new.PublishUrl {
		return true
	}
	if old.Fps != new.Fps {
		return true
	}
	if old.Height != new.Width {
		return true
	}
	if old.Width != new.Width {
		return true
	}
	return false
}

func (c *MediaController) markForReconcile(cameraId string) {
	for _, id := range c.needConcilation {
		if id == cameraId {
			return
		}
	}
	c.needConcilation = append(
		c.needConcilation,
		cameraId)
}

func (c *MediaController) startOrRestart(cameraId string) error {
	p, found := c.running[cameraId]
	if found {
		if err := c.stop(cameraId); err != nil {
			logger.SError("error stopping stream",
				zap.String("cameraId", cameraId),
				zap.Error(err))
			return err
		}
		delete(c.running, cameraId)
		logger.SDebug("stopped stream",
			zap.String("cameraId", cameraId))
	}

	s, found := c.ffmpegStreams[cameraId]
	if !found {
		logger.SDebug("stream not found",
			zap.String("cameraId", cameraId))
		return nil
	}

	p = &Process{
		cameraId: cameraId,
		configs:  &s,
	}
	go func(cameraId string) {
		if err := c.mediaService.StartTranscodingStream(context.Background(), p); err != nil {
			if c.Exists(cameraId) {
				c.mu.Lock()
				c.markForReconcile(cameraId)
				delete(c.running, cameraId)
				c.mu.Unlock()
			}

			logger.SDebug("error starting stream",
				zap.String("cameraId", cameraId))
			return
		}
		logger.SDebug("exited stream",
			zap.String("cameraId", cameraId))
	}(cameraId)

	c.running[cameraId] = p
	return nil
}

func (c *MediaController) stop(cameraId string) error {
	p, found := c.running[cameraId]
	if !found {
		logger.SDebug("stream not found",
			zap.String("cameraId", cameraId))
		return nil
	}

	if err := c.mediaService.EndTranscodingStream(context.Background(), p); err != nil {
		logger.SDebug("error stopping stream",
			zap.String("cameraId", cameraId))
		return err
	}

	delete(c.running, cameraId)
	return nil
}

type ProcessorController struct {
	mu              sync.Mutex
	configs         *configs.OpenGateConfigs
	settings        []byte
	updatedSettings []byte
	running         bool
	mediaService    MediaServiceInterface
	proc            *OpenGateProcess
}

func NewProcessorController(
	configs *configs.OpenGateConfigs,
	mediaService MediaServiceInterface) *ProcessorController {
	return &ProcessorController{
		configs:      configs,
		running:      false,
		mediaService: mediaService,
	}
}

func (c *ProcessorController) Reconcile(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logger.SDebug("processor controller context cancelled")
			if err := c.shutdown(ctx); err != nil {
				logger.SError("error shutting down processor",
					zap.Error(err))
				return err
			}
			logger.SDebug("processor controller cleaned up")
			return nil
		default:
			if err := c.reconcile(ctx); err != nil {
				logger.SError("error reconciling processor",
					zap.Error(err))
				return err
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
	}
}

func (c *ProcessorController) Updates(settings []byte) {
	c.mu.Lock()
	c.updatedSettings = settings
	c.mu.Unlock()
}

func (c *ProcessorController) reconcile(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	updated := c.compareSettings()
	if updated {
		c.settings = c.updatedSettings
	}
	switch c.running {
	case true:
		if updated {
			logger.SInfo("processor is running, settings is updated, restarting")
			if err := c.startOrRestart(ctx); err != nil {
				logger.SError("error restarting processor",
					zap.Error(err))
				return err
			}
			break
		}
	case false:
		if len(c.settings) > 0 {
			logger.SInfo("processor is not running, starting")
			if err := c.startOrRestart(ctx); err != nil {
				logger.SError("error starting processor",
					zap.Error(err))
				return err
			}
		}
	}
	return nil
}

func (c *ProcessorController) shutdown(ctx context.Context) error {
	if c.proc != nil {
		if c.running {
			logger.SInfo("shutting down processor")
			if err := c.mediaService.ComposeDownOpenGate(ctx, c.proc); err != nil {
				logger.SError("error shutting down processor",
					zap.Error(err))
				return err
			}
			c.running = false
		}
	}
	return nil
}

func (c *ProcessorController) startOrRestart(ctx context.Context) error {
	c.proc = &OpenGateProcess{
		configs:  c.configs,
		settings: c.updatedSettings,
	}
	if c.running {
		if err := c.mediaService.ComposeRestartOpenGate(ctx, c.proc); err != nil {
			logger.SError("error restarting processor",
				zap.Error(err))
			c.running = false
			return err
		}
		return nil
	}
	if err := c.mediaService.ComposeUpOpenGate(ctx, c.proc); err != nil {
		logger.SError("error starting processor",
			zap.Error(err))
		c.running = false
		return err
	}
	c.running = true
	go c.watchStartCmd(c.proc.proc)

	logger.SInfo("processor started",
		zap.String("settings", string(c.updatedSettings)))
	return nil
}

func (c *ProcessorController) watchStartCmd(cmd *exec.Cmd) {
	for {
		if cmd.Process != nil {
			if cmd.ProcessState != nil {
				if cmd.ProcessState.Exited() {
					logger.SInfo("start command exited successfully")
					return
				} else {
					logger.SDebug("start command is running")
				}
			}
		}
		for range time.After(time.Second * 1) {
			continue
		}
	}
}

func (c *ProcessorController) compareSettings() (updated bool) {
	if len(c.updatedSettings) == 0 {
		return false
	}
	if string(c.settings) == string(c.updatedSettings) {
		return false
	}
	return true
}

func (c *ProcessorController) PrePullImages(ctx context.Context) error {
	logger.SDebug("pre-pulling images", zap.Reflect("images", c.configs.PrepullImages))
	images := c.configs.PrepullImages
	if err := c.mediaService.DockerPullImages(ctx, images); err != nil {
		logger.SError("error pre-pulling images",
			zap.Error(err))
		return err
	}
	logger.SInfo("images pulled successfully")
	return nil
}
