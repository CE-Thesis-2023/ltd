package reconciler

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Reconciler struct {
	cameras             map[string]*db.Camera
	controlPlaneService *service.ControlPlaneService
	deviceInfo          *configs.DeviceInfoConfigs
	commandService      *service.CommandService
	mu                  sync.Mutex
}

func NewReconciler(controlPlaneService *service.ControlPlaneService, deviceInfo *configs.DeviceInfoConfigs, commandService *service.CommandService) *Reconciler {
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
	return &Reconciler{
		cameras:             make(map[string]*db.Camera),
		controlPlaneService: controlPlaneService,
		deviceInfo:          deviceInfo,
		commandService:      commandService,
	}
}

func (c *Reconciler) Run(ctx context.Context) {
	logger.SInfo("reconciler loop Enabled")
	if err := c.initApplication(ctx); err != nil {
		logger.SError("reconciler loop initialize application failed",
			zap.Error(err))
		return
	}
	for {
		if err := c.reconcile(ctx); err != nil {
			logger.SError("reconciler loop reconcile failed",
				zap.Error(err))
			return
		}
		time.Sleep(5 * time.Second)
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
	if len(cameras) == 0 {
		logger.SInfo("no cameras assigned, nothing to do")
		return nil
	}
	cameraIds := make([]string, 0, len(cameras))
	for _, camera := range cameras {
		cameraIds = append(cameraIds, camera.CameraId)
	}
	logger.SDebug("assigned cameras",
		zap.Reflect("cameras", cameraIds))

	openGateConfigurations, err := c.controlPlaneService.GetOpenGateConfigurations(ctx, &web.GetTranscoderOpenGateConfigurationRequest{
		TranscoderId: c.deviceInfo.DeviceId,
	})
	if err != nil {
		logger.SError("failed to get open gate configurations",
			zap.Error(err))
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(openGateConfigurations.Base64)
	if err != nil {
		logger.SError("failed to decode base64 open gate configurations",
			zap.Error(err))
		return err
	}

	var jsonConfig map[string]interface{}
	yaml.Unmarshal(decoded, &jsonConfig)
	logger.SDebug("OpenGate configurations",
		zap.Reflect("configurations", jsonConfig))

	cameraConfigurations, err := c.controlPlaneService.GetCameraStreamSettings(ctx, &web.GetStreamConfigurationsRequest{
		CameraId: cameraIds,
	})
	if err != nil {
		logger.SError("failed to get camera stream settings",
			zap.Error(err))
		return err
	}
	logger.SDebug("camera stream settings",
		zap.Reflect("settings", cameraConfigurations))

	logger.SInfo("reconcile completed")
	return nil
}
