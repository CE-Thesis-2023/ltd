package eventsapi

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/cache"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/concurrent"

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/ristretto"
	"github.com/eclipse/paho.golang/paho"
	"github.com/mitchellh/mapstructure"
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
			var info events.CommandRetrieveDeviceInfo
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
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
			var info events.CommandAddCameraInfo
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
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
			var info events.CommandStartStreamInfo
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
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
			var info events.CommandEndStreamInfo
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: Command_EndStream",
					zap.String("error", "info not type CommmandEndStreamInfo"))
				return
			}
			if err := biz.EndStream(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.EndStream", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_EndStream success")
		case events.Command_StartFfmpegStream:
			var info events.CommandStartStreamInfo
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: Command_StartFfmpegStream",
					zap.String("error", "info not type CameraStartStreamInfo"))
				return
			}
			if err := biz.StartFfmpegStream(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.StartFfmpegStream", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_StartFfmpegStream success")
		case events.Command_EndFfmpegStream:
			var info events.CommandEndStreamInfo
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: Command_EndFfmpegStream",
					zap.String("error", "info not type CommmandEndStreamInfo"))
				return
			}
			if err := biz.EndFfmpegStream(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.EndFfmpegStream", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_EndFfmpegStream success")
		case events.Command_GetStreamChannels:
			var info events.CommandRetrieveStreamChannels
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: Command_GetStreamChannels",
					zap.String("error", "info not type CommandRetrieveStreamChannels"))
				return
			}
			if err := biz.StreamChannels(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.StreamChannels", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_GetStreamChannels success")
		case events.Command_GetStreamStatus:
			var info events.CommandGetStreamStatusRequest
			if err := mapstructure.Decode(&msg.Info, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: Command_GetStreamStatus",
					zap.String("error",
						"info not type CommandGetStreamStatusRequest"))
				return
			}
			if err := biz.StreamStatus(ctx, &info); err != nil {
				logger.SError("ReceiveRemoteCommands: biz.StreamStatus", zap.Error(err))
				return
			}
			logger.SInfo("ReceiveRemoteCommands: Command_GetStreamStatus success")
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
