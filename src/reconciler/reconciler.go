package reconciler

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/service"
	"go.uber.org/zap"
)

type Reconciler struct {
	cameras         map[string]web.TranscoderStreamConfiguration
	openGateConfigs string

	controlPlaneService *service.ControlPlaneService
	deviceInfo          *configs.DeviceInfoConfigs
	commandService      *service.CommandService
	mediaService        *service.MediaController

	updatedOpenGateConfigs string
	updatedCameras         map[string]web.TranscoderStreamConfiguration

	mu sync.Mutex
}

func NewReconciler(
	controlPlaneService *service.ControlPlaneService,
	deviceInfo *configs.DeviceInfoConfigs,
	commandService *service.CommandService,
	mediaService *service.MediaController) *Reconciler {
	if controlPlaneService == nil {
		logger.SFatal("control plane service is nil",
			zap.String("error", "control plane service is nil"))
	}
	if deviceInfo == nil {
		logger.SFatal("device info is nil",
			zap.String("error", "device info is nil"))
	}
	if commandService == nil {
		logger.SFatal("command service is nil",
			zap.String("error", "command service is nil"))
	}
	if mediaService == nil {
		logger.SFatal("media service is nil",
			zap.String("error", "media service is nil"))
	}
	return &Reconciler{
		cameras:             make(map[string]web.TranscoderStreamConfiguration),
		controlPlaneService: controlPlaneService,
		deviceInfo:          deviceInfo,
		commandService:      commandService,
		mediaService:        mediaService,
	}
}

func (c *Reconciler) Run(ctx context.Context) {
	logger.SInfo("reconciler loop Enabled")
	if err := c.init(ctx); err != nil {
		logger.SError("reconciler loop initialize application failed",
			zap.Error(err))
		return
	}

	go func() {
		if err := c.mediaService.Reconcile(ctx); err != nil {
			logger.SFatal("media controller reconcile failed",
				zap.Error(err))
		}
	}()

	for {
		if err := c.reconcile(ctx); err != nil {
			logger.SError("reconciler loop reconcile failed",
				zap.Error(err))
			return
		}

		select {
		case <-time.After(2 * time.Second):
			continue
		case <-ctx.Done():
			logger.SInfo("reconciler loop shutdown requested")
			return
		}
	}
}

func (c *Reconciler) init(ctx context.Context) error {
	err := c.controlPlaneService.
		RegisterDevice(ctx, &service.RegistrationRequest{
			DeviceId: c.
				deviceInfo.
				DeviceId,
		})
	switch err {
	case custerror.ErrorAlreadyExists:
		logger.SInfo("device already registered",
			zap.String("id", c.deviceInfo.DeviceId))
	case nil:
		logger.SInfo("device registered successfully",
			zap.String("id", c.deviceInfo.DeviceId))
	default:
		logger.SError("unable to register the device",
			zap.Error(err))
		return err
	}

	mqttEndpoints, err := c.controlPlaneService.GetMQTTEndpoints(ctx, &web.GetMQTTEventEndpointRequest{
		TranscoderId: c.deviceInfo.DeviceId,
	})
	if err != nil {
		logger.SError("failed to get MQTT endpoints",
			zap.Error(err))
		return err
	}
	logger.SDebug("MQTT endpoints",
		zap.Reflect("endpoints", mqttEndpoints))
	return nil
}

func (c *Reconciler) reconcile(ctx context.Context) error {
	logger.SInfo("reconcile Enabled")

	if err := c.pullLatestConfigurations(ctx); err != nil {
		logger.SError("failed to pull latest configurations",
			zap.Error(err))
		return err
	}

	if err := c.reconcileFFmpegStreams(); err != nil {
		logger.SError("failed to reconcile FFmpeg streams",
			zap.Error(err))
		return err
	}

	logger.SInfo("reconcile completed")
	return nil
}

func (c *Reconciler) pullLatestConfigurations(ctx context.Context) error {
	logger.SInfo("pulling latest configurations",
		zap.String("transcoderId", c.deviceInfo.DeviceId))

	if err := c.pullOpenGateConfiguration(ctx); err != nil {
		logger.SError("failed to pull open gate configuration",
			zap.Error(err))
		return err
	}
	if err := c.pullStreamConfigurations(ctx); err != nil {
		logger.SError("failed to pull stream configurations",
			zap.Error(err))
		return err
	}

	logger.SInfo("pulling latest configurations completed")
	return nil
}

func (c *Reconciler) pullOpenGateConfiguration(ctx context.Context) error {
	openGateResp, err := c.controlPlaneService.GetOpenGateConfigurations(ctx, &web.GetTranscoderOpenGateConfigurationRequest{
		TranscoderId: c.deviceInfo.DeviceId,
	})
	if err != nil {
		return err
	}
	decoded, err := base64.
		StdEncoding.
		DecodeString(openGateResp.Base64)
	if err != nil {
		return err
	}
	c.updatedOpenGateConfigs = string(decoded)
	return nil
}

func (c *Reconciler) pullStreamConfigurations(ctx context.Context) error {
	logger.SInfo("pulling stream configurations",
		zap.String("transcoderId", c.deviceInfo.DeviceId))

	assignedResp, err := c.controlPlaneService.GetAssignedDevices(ctx, &service.GetAssignedDevicesRequest{
		DeviceId: c.
			deviceInfo.
			DeviceId,
	})
	if err != nil {
		return err
	}
	cameras := assignedResp.Cameras
	cameraIds := make([]string, 0, len(cameras))
	for _, camera := range cameras {
		cameraIds = append(cameraIds, camera.CameraId)
	}
	c.updatedCameras = make(map[string]web.TranscoderStreamConfiguration)
	if len(cameraIds) > 0 {
		cameraConfigurations, err := c.controlPlaneService.GetCameraStreamSettings(ctx, &web.GetStreamConfigurationsRequest{
			CameraId: cameraIds,
		})
		if err != nil {
			return err
		}
		configs := cameraConfigurations.StreamConfigurations
		for _, config := range configs {
			c.updatedCameras[config.CameraId] = config
		}
	}
	return nil
}

func (c *Reconciler) reconcileFFmpegStreams() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for cameraId, newConfig := range c.updatedCameras {
		if _, ok := c.cameras[cameraId]; !ok {
			logger.SInfo("new camera stream configuration",
				zap.String("cameraId", cameraId))
			updated, err := c.mediaService.Register(newConfig)
			if err != nil {
				logger.SError("failed to register camera stream configuration",
					zap.String("cameraId", cameraId),
					zap.Error(err))
				return err
			}
			if updated {
				logger.SInfo("camera stream configuration updated",
					zap.String("cameraId", cameraId))
			}
		}
	}
	for cameraId := range c.cameras {
		if _, ok := c.updatedCameras[cameraId]; !ok {
			logger.SInfo("camera stream configuration removed",
				zap.String("cameraId", cameraId))
			c.mediaService.Deregister(cameraId)
			logger.SDebug("camera stream configuration removed")
		}
	}
	c.cameras = c.updatedCameras
	return nil
}

func (c *Reconciler) reconcileOpenGate() error {
	if c.openGateConfigs != c.updatedOpenGateConfigs {
		logger.SInfo("OpenGate configuration updated")
		c.openGateConfigs = c.updatedOpenGateConfigs
	}

	return nil
}
