package reconciler

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Reconciler struct {
	cameras               map[string]*db.Camera
	controlPlaneService   *service.ControlPlaneService
	deviceInfo            *configs.DeviceInfoConfigs
	commandService        *service.CommandService
	shouldRebuildOpenGate bool
	mu                    sync.Mutex
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

func (c *Reconciler) matchCameras(
	updatedCameras []db.Camera,
	onRemove func(c *db.Camera) error,
	checkUpdates func(old *db.Camera, updated *db.Camera) (bool, error),
	onAddition func(c *db.Camera) error,
	rebuildOpenGateConfiguration func(updatedCameras []db.Camera) error) error {
	mappedUpdatedDevices := make(map[string]*db.Camera)
	for _, camera := range updatedCameras {
		mappedUpdatedDevices[camera.CameraId] = &camera
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.shouldRebuildOpenGate = false
	merged := make(map[string]*db.Camera)

	for cameraId, new := range mappedUpdatedDevices {
		if old, found := c.cameras[cameraId]; found {
			if updated, err := checkUpdates(old, new); err != nil {
				merged[cameraId] = old
			} else {
				merged[cameraId] = new
				if updated {
					c.shouldRebuildOpenGate = true
				}
			}
		} else {
			if err := onAddition(new); err != nil {
				logger.SError("failed to add camera",
					zap.Error(err))
			} else {
				merged[cameraId] = new
				c.shouldRebuildOpenGate = true
			}
		}
	}
	for cameraId, old := range c.cameras {
		if _, found := mappedUpdatedDevices[cameraId]; !found {
			if err := onRemove(old); err != nil {
				merged[cameraId] = old
			} else {
				c.shouldRebuildOpenGate = true
			}
		}
	}

	if c.shouldRebuildOpenGate {
		if err := rebuildOpenGateConfiguration(updatedCameras); err != nil {
			logger.SError("failed to rebuild open gate configuration",
				zap.Error(err))
			return err
		}
	}

	c.cameras = merged
	return nil
}

func (c *Reconciler) rebuildOpenGateConfiguration(updatedCameras []db.Camera) error {
	logger.SInfo("rebuild open gate configuration")

	return nil
}

func (c *Reconciler) onAddition(camera *db.Camera) error {
	logger.SInfo("camera added",
		zap.String("id", camera.CameraId))

	// if err := c.commandService.StartFfmpegStream(
	// 	context.Background(),
	// 	camera,
	// 	&events.CommandStartStreamInfo{
	// 		CameraId: camera.CameraId,
	// 	}); err != nil {
	// 	logger.SError("failed to start ffmpeg stream",
	// 		zap.Error(err))
	// 	return err
	// }
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

func (c *Reconciler) checkUpdates(old *db.Camera, updated *db.Camera) (changed bool, err error) {
	logger.SInfo("camera updated",
		zap.String("id", updated.CameraId))
	changed = false
	if old.Ip != updated.Ip {
		logger.SInfo("camera ip updated",
			zap.String("id", updated.CameraId),
			zap.String("old_ip", old.Ip),
			zap.String("new_ip", updated.Ip))
		changed = true
	}
	if old.Enabled != updated.Enabled {
		logger.SInfo("camera stream status updated",
			zap.String("id", updated.CameraId),
			zap.Bool("old_Enabled", old.Enabled),
			zap.Bool("new_Enabled", updated.Enabled))
		changed = true
	}
	if changed {
		if old.Enabled {
			if err := c.commandService.EndFfmpegStream(
				context.Background(),
				old,
				&events.CommandEndStreamInfo{
					CameraId: old.CameraId,
				}); err != nil {
				logger.SError("failed to end ffmpeg stream",
					zap.Error(err))
				if err != custerror.ErrorNotFound {
					return changed, err
				}
			}
		}
		if updated.Enabled {
			// if err := c.commandService.StartFfmpegStream(
			// 	context.Background(),
			// 	updated,
			// 	&events.CommandStartStreamInfo{
			// 		CameraId: updated.CameraId,
			// 	}); err != nil {
			// 	logger.SError("failed to start ffmpeg stream",
			// 		zap.Error(err))
			// 	if err != custerror.ErrorAlreadyExists {
			// 		return changed, err
			// 	}
			// 	return changed, err
			// }
		}
	}
	return changed, err
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
