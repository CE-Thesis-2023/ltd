package handlers

import (
	"context"

	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

func CommandHandlers(ctx context.Context, msg *events.CommandRequest) (*events.CommandResponse, error) {
	biz := service.GetCommandService()
	switch msg.CommandType {
	case events.Command_GetDeviceInfo:
		var info events.CommandRetrieveDeviceInfo
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_GetDeviceInfo",
				zap.String("error", "info not type CommandRetrieveDeviceInfo"))
			return nil, err
		}
		if err := biz.DeviceInfo(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.DeviceInfo", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_GetDeviceInfo success")

	case events.Command_AddCamera:
		var info events.CommandAddCameraInfo
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_AddCamera",
				zap.String("error", "info not type CommandAddCameraInfo"))
			return nil, err
		}
		if err := biz.AddCamera(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.AddCamera", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_AddCamera success")
	case events.Command_StartFfmpegStream:
		var info events.CommandStartStreamInfo
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_StartFfmpegStream",
				zap.String("error", "info not type CameraStartStreamInfo"))
			return nil, err
		}
		if err := biz.StartFfmpegStream(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.StartFfmpegStream", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_StartFfmpegStream success")
	case events.Command_EndFfmpegStream:
		var info events.CommandEndStreamInfo
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_EndFfmpegStream",
				zap.String("error", "info not type CommmandEndStreamInfo"))
			return nil, err
		}
		if err := biz.EndFfmpegStream(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.EndFfmpegStream", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_EndFfmpegStream success")
	case events.Command_GetStreamChannels:
		var info events.CommandRetrieveStreamChannels
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_GetStreamChannels",
				zap.String("error", "info not type CommandRetrieveStreamChannels"))
			return nil, err
		}
		if err := biz.StreamChannels(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.StreamChannels", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_GetStreamChannels success")
	case events.Command_GetStreamStatus:
		var info events.CommandGetStreamStatusRequest
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_GetStreamStatus",
				zap.String("error",
					"info not type CommandGetStreamStatusRequest"))
			return nil, err
		}
		if err := biz.StreamStatus(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.StreamStatus", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_GetStreamStatus success")
	case events.Command_DeleteCamera:
		var info events.CommandDeleteCameraRequest
		if err := mapstructure.Decode(&msg.Info, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: Command_DeleteCamera",
				zap.String("error",
					"info not type CommandDeleteCameraRequest"))
			return nil, err
		}
		if err := biz.DeleteCamera(ctx, &info); err != nil {
			logger.SError("ReceiveRemoteCommands: biz.DeleteCamera", zap.Error(err))
			return nil, err
		}
		logger.SInfo("ReceiveRemoteCommands: Command_DeleteCamera success")
	default:
		logger.SError("ReceiveRemoteCommands: unknown command type",
			zap.String("type", string(msg.CommandType)),
			zap.String("do", "skipping"))
	}
	return nil, nil
}
