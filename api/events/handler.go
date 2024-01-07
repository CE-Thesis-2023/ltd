package eventsapi

import (
	"context"
	"labs/local-transcoder/biz/service"
	"labs/local-transcoder/internal/cache"
	"labs/local-transcoder/internal/logger"
	"labs/local-transcoder/models/events"
	"time"

	"labs/local-transcoder/internal/concurrent"

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/ristretto"
	"github.com/eclipse/paho.golang/paho"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type StandardEventHandler struct {
	pool  *ants.Pool
	cache *ristretto.Cache
}

func NewStandardEventHandler() *StandardEventHandler {
	return &StandardEventHandler{
		pool:  custcon.New(100),
		cache: cache.Cache(),
	}
}

func (h *StandardEventHandler) ReceiveRemoteCommands(p *paho.Publish) error {
	logger.SDebug("ReceiveRemoteCommands", zap.String("message", string(p.Payload)))

	var msg events.CommandRequest
	if err := sonic.Unmarshal(p.Payload, &msg); err != nil {
		logger.SError("ReceiveRemoteCommands: message parsing failed", zap.Error(err))
		return err
	}

	h.pool.Submit(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer func() {
			if ctx.Err() != nil {
				logger.SDebug("ReceiveRemoteCommands: context exceeded")
			}
			cancel()
		}()

		biz := service.GetCommandService()
		switch msg.CommandType {
		case events.Command_GetDeviceInfo:
			info, ok := msg.Info.(events.CommandRetrieveDeviceInfo)
			if !ok {
				logger.SError("ReceiveRemoteCommands: Command_GetDeviceInfo",
					zap.String("error", "info not type CommandRetrieveDeviceInfo"))
				return
			}
			if err := biz.DeviceInfo(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.DeviceInfo", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_GetDeviceInfo success")
		case events.Command_AddCamera:
			info, ok := msg.Info.(events.CommandAddCameraInfo)
			if !ok {
				logger.SError("ReceiveRemoteCommands: Command_AddCamera",
					zap.String("error", "info not type CommandAddCameraInfo"))
				return
			}
			if err := biz.AddCamera(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.AddCamera", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_AddCamera success")
		case events.Command_StartStream:
			info, ok := msg.Info.(events.CommandStartStreamInfo)
			if !ok {
				logger.SError("ReceiveRemoteCommands: Command_StartStream",
					zap.String("error", "info not type CameraStartStreamInfo"))
				return
			}
			if err := biz.StartStream(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.StartStream", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_StartStream success")
		case events.Command_EndStream:
			info, ok := msg.Info.(events.CommandEndStreamInfo)
			if !ok {
				logger.SError("ReceiveRemoteCommands: Command_EndStream",
					zap.String("error", "info not type CommmandEndStreamInfo"))
				return
			}
			if err := biz.EndStream(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.EndStream", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_EndStream success")
		default:
			logger.SError("ReceiveRemoteCommands: unknown command type",
				zap.String("type", string(msg.CommandType)),
				zap.String("do", "skipping"))
		}
	})

	logger.SDebug("ReceiveRemoteCommands: goroutine assigned")
	return nil
}

func (h *StandardEventHandler) ReceiveRemoteMovementControl(p *paho.Publish) error {
	logger.SDebug("ReceiveRemoteMovementControl", zap.String("message", string(p.Payload)))

	var msg events.PtzCtrlRequest
	if err := sonic.Unmarshal(p.Payload, &msg); err != nil {
		logger.SError("ReceiveRemoteMovementControl: message parsing failed", zap.Error(err))
		return err
	}

	dur := time.Second * (time.Duration(*msg.StopAfterSeconds) + 2)
	ctx, cancel := context.WithTimeout(context.Background(), dur)
	defer func() {
		if ctx.Err() != nil {
			logger.SDebug("ReceiveRemoteMovementControl: context exceeded")
		}
		cancel()
	}()
	if err := service.GetCommandService().PtzCtrl(ctx, &msg); err != nil {
		logger.SError("ReceiveRemoteMovementControl: CommandService.PtzCtrl", zap.Error(err))
		return err
	}

	logger.SDebug("ReceiveRemoteMovementControl: success")
	return nil
}
