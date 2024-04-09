package reconciler

import (
	"context"
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
	cameras             map[string]web.TranscoderStreamConfiguration
	controlPlaneService *service.ControlPlaneService
	deviceInfo          *configs.DeviceInfoConfigs
	commandService      *service.CommandService
	mediaService        *service.MediaController
	mu                  sync.Mutex
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
	if err := c.initApplication(ctx); err != nil {
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

func (c *Reconciler) initApplication(ctx context.Context) error {
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
	resp, err := c.controlPlaneService.GetAssignedDevices(ctx, &service.GetAssignedDevicesRequest{
		DeviceId: c.
			deviceInfo.
			DeviceId,
	})
	if err != nil {
		return err
	}
	cameras := resp.Cameras

	cameraIds := make([]string, 0, len(cameras))
	for _, camera := range cameras {
		cameraIds = append(cameraIds, camera.CameraId)
	}
	logger.SDebug("assigned cameras",
		zap.Reflect("cameras", cameraIds))

	mapped := make(map[string]web.TranscoderStreamConfiguration)
	if len(cameras) > 0 {
		_, err = c.controlPlaneService.GetOpenGateConfigurations(ctx, &web.GetTranscoderOpenGateConfigurationRequest{
			TranscoderId: c.deviceInfo.DeviceId,
		})
		if err != nil {
			logger.SError("failed to get open gate configurations",
				zap.Error(err))
			return err
		}

		cameraConfigurations, err := c.controlPlaneService.GetCameraStreamSettings(ctx, &web.GetStreamConfigurationsRequest{
			CameraId: cameraIds,
		})
		if err != nil {
			logger.SError("failed to get camera stream settings",
				zap.Error(err))
			return err
		}
		configs := cameraConfigurations.StreamConfigurations
		for _, config := range configs {
			mapped[config.CameraId] = config
		}
	}

	c.reconcileFFmpegStreams(mapped)

	logger.SInfo("reconcile completed")
	return nil
}

func (c *Reconciler) reconcileFFmpegStreams(newCameraConfigs map[string]web.TranscoderStreamConfiguration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for cameraId, newConfig := range newCameraConfigs {
		if _, ok := c.cameras[cameraId]; !ok {
			logger.SInfo("new camera stream configuration",
				zap.String("cameraId", cameraId))
			updated, err := c.mediaService.Register(newConfig)
			if err != nil {
				logger.SError("failed to register camera stream configuration",
					zap.String("cameraId", cameraId),
					zap.Error(err))
				continue
			}
			if updated {
				logger.SInfo("camera stream configuration updated",
					zap.String("cameraId", cameraId))
			}
		}
	}
	for cameraId := range c.cameras {
		if _, ok := newCameraConfigs[cameraId]; !ok {
			logger.SInfo("camera stream configuration removed",
				zap.String("cameraId", cameraId))
			c.mediaService.Deregister(cameraId)
			logger.SDebug("camera stream configuration removed")
		}
	}
	c.cameras = newCameraConfigs
}
