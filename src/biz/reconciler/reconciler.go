package reconciler

import (
	"context"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Reconciler struct {
	cameras             map[string]*db.Camera
	controlPlaneService *service.ControlPlaneService
	deviceInfo          *configs.DeviceInfoConfigs
	mu                  sync.Mutex
}

func NewReconciler(controlPlaneService *service.ControlPlaneService, deviceInfo *configs.DeviceInfoConfigs) *Reconciler {
	if controlPlaneService == nil {
		logger.SFatal("control plane service is nil",
			zap.String("error", "control plane service is nil"))
	}
	if deviceInfo == nil {
		logger.SFatal("device info is nil",
			zap.String("error", "device info is nil"))
	}
	return &Reconciler{
		cameras:             make(map[string]*db.Camera),
		controlPlaneService: controlPlaneService,
		deviceInfo:          deviceInfo,
	}
}

// Runs an event loop that listens for events and processes them
func (c *Reconciler) Run(ctx context.Context) {
	logger.SInfo("reconciler loop started")
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
		time.Sleep(10 * time.Second)
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
	logger.SInfo("reconcile started")
	resp, err := c.controlPlaneService.GetAssignedDevices(ctx, &service.GetAssignedDevicesRequest{
		DeviceId: c.
			deviceInfo.
			DeviceId,
	})
	if err != nil {
		return err
	}
	devices := resp.Cameras
	if err := c.matchCameras(devices,
		c.onRemove,
		c.checkUpdates,
		c.onAddition); err != nil {
		return err
	}
	logger.SInfo("reconcile completed")
	return nil
}

func (c *Reconciler) matchCameras(updatedCameras []db.Camera, onRemove func(c *db.Camera) error, checkUpdates func(c *db.Camera) error, onAddition func(c *db.Camera)) error {
	mappedUpdatedDevices := make(map[string]*db.Camera)
	for _, camera := range updatedCameras {
		mappedUpdatedDevices[camera.CameraId] = &camera
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for cameraId, camera := range mappedUpdatedDevices {
		if _, found := c.cameras[cameraId]; !found {
			onAddition(camera)
		} else {
			if err := checkUpdates(camera); err != nil {
				return err
			}
			delete(c.cameras, cameraId)
		}
	}
	for _, camera := range c.cameras {
		if err := onRemove(camera); err != nil {
			return err
		}
	}
	c.cameras = mappedUpdatedDevices
	return nil
}

func (c *Reconciler) onAddition(camera *db.Camera) {
	logger.SInfo("camera added",
		zap.String("id", camera.CameraId))
}

func (c *Reconciler) onRemove(camera *db.Camera) error {
	logger.SInfo("camera removed",
		zap.String("id", camera.CameraId))
	return nil
}

func (c *Reconciler) checkUpdates(camera *db.Camera) error {
	logger.SInfo("camera updated",
		zap.String("id", camera.CameraId))
	return nil
}

func (c *Reconciler) ProcessInputEvent(ctx context.Context, msg *events.CommandRequest) error {
	biz := service.GetCommandService()
	switch msg.CommandType {
	case events.Command_GetDeviceInfo:
		var info events.CommandRetrieveDeviceInfo
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("get device info command failed",
				zap.String("error", "info not type CommandRetrieveDeviceInfo"))
			return err
		}
		camera, found := c.cameras[info.CameraId]
		if !found {
			logger.SError("get device info command failed",
				zap.String("error", "camera not found"))
			return custerror.ErrorNotFound
		}
		resp, err := biz.DeviceInfo(ctx, camera)
		if err != nil {
			logger.SError("get device info command failed", zap.Error(err))
			return err
		}
		logger.SInfo("get device info command success",
			zap.Any("response", resp))
	default:
		logger.SError("unknown command type",
			zap.String("type", string(msg.CommandType)),
			zap.String("do", "skipping"))
	}
	return nil
}
