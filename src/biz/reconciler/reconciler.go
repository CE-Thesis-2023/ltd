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
	if len(devices) == 0 {
		logger.SInfo("no cameras assigned, nothing to do")
		return nil
	}
	if err := c.matchCameras(devices,
		c.onRemove,
		c.checkUpdates,
		c.onAddition); err != nil {
		return err
	}
	logger.SInfo("reconcile completed")
	return nil
}

func (c *Reconciler) matchCameras(updatedCameras []db.Camera, onRemove func(c *db.Camera) error, checkUpdates func(old *db.Camera, updated *db.Camera) error, onAddition func(c *db.Camera) error) error {
	mappedUpdatedDevices := make(map[string]*db.Camera)
	for _, camera := range updatedCameras {
		mappedUpdatedDevices[camera.CameraId] = &camera
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	merged := make(map[string]*db.Camera)
	for cameraId, new := range mappedUpdatedDevices {
		if old, found := c.cameras[cameraId]; found {
			if err := checkUpdates(old, new); err != nil {
				merged[cameraId] = old
			} else {
				merged[cameraId] = new
			}
		} else {
			if err := onAddition(new); err != nil {
				logger.SError("failed to add camera",
					zap.Error(err))
			} else {
				merged[cameraId] = new
			}
		}
	}
	for cameraId, old := range c.cameras {
		if _, found := mappedUpdatedDevices[cameraId]; !found {
			if err := onRemove(old); err != nil {
				merged[cameraId] = old
			}
		}
	}
	logger.SDebug("cameras merged",
		zap.Reflect("old", c.cameras),
		zap.Reflect("new", merged))
	c.cameras = merged
	return nil
}

func (c *Reconciler) onAddition(camera *db.Camera) error {
	logger.SInfo("camera added",
		zap.String("id", camera.CameraId))

	if err := c.commandService.StartFfmpegStream(
		context.Background(),
		camera,
		&events.CommandStartStreamInfo{
			CameraId: camera.CameraId,
		}); err != nil {
		logger.SError("failed to start ffmpeg stream",
			zap.Error(err))
		return err
	}
	return nil
}

func (c *Reconciler) onRemove(camera *db.Camera) error {
	logger.SInfo("camera removed",
		zap.String("id", camera.CameraId))
	if err := c.commandService.EndFfmpegStream(
		context.Background(),
		camera,
		&events.CommandEndStreamInfo{
			CameraId: camera.CameraId,
		}); err != nil {
		logger.SError("failed to end ffmpeg stream",
			zap.Error(err))
		return err
	}
	return nil
}

func (c *Reconciler) checkUpdates(old *db.Camera, updated *db.Camera) error {
	logger.SInfo("camera updated",
		zap.String("id", updated.CameraId))
	changed := false
	if old.Ip != updated.Ip {
		logger.SInfo("camera ip updated",
			zap.String("id", updated.CameraId),
			zap.String("old_ip", old.Ip),
			zap.String("new_ip", updated.Ip))
		changed = true
	}
	if old.Started != updated.Started {
		logger.SInfo("camera stream status updated",
			zap.String("id", updated.CameraId),
			zap.Bool("old_started", old.Started),
			zap.Bool("new_started", updated.Started))
		changed = true
	}
	if changed {
		if old.Started {
			if err := c.commandService.EndFfmpegStream(
				context.Background(),
				old,
				&events.CommandEndStreamInfo{
					CameraId: old.CameraId,
				}); err != nil {
				logger.SError("failed to end ffmpeg stream",
					zap.Error(err))
				if err != custerror.ErrorNotFound {
					return err
				}
			}
		}
		if updated.Started {
			if err := c.commandService.StartFfmpegStream(
				context.Background(),
				updated,
				&events.CommandStartStreamInfo{
					CameraId: updated.CameraId,
				}); err != nil {
				logger.SError("failed to start ffmpeg stream",
					zap.Error(err))
				if err != custerror.ErrorAlreadyExists {
					return err
				}
				return err
			}
		}
	}
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
			zap.Reflect("response", resp))
	default:
		logger.SError("unknown command type",
			zap.String("type", string(msg.CommandType)),
			zap.String("do", "skipping"))
	}
	return nil
}
