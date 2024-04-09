package service

import (
	"context"
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
		select {
		case <-ctx.Done():
			logger.SDebug("media controller context cancelled",
				zap.Error(ctx.Err()))
			c.cleanup()
			return ctx.Err()
		default:
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
	configs *configs.OpenGateConfigs
}

func NewProcessorController(configs *configs.OpenGateConfigs) *ProcessorController {
	return &ProcessorController{
		configs: configs,
	}
}

func (c *ProcessorController) Reconcile(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logger.SDebug("processor controller context cancelled",
				zap.Error(ctx.Err()))
			return ctx.Err()
		default:
			time.Sleep(200 * time.Millisecond)
			continue
		}
	}
}
