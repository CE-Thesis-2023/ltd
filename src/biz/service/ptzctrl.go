package service

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"time"

	"go.uber.org/zap"
)

func (s *CommandService) PtzCtrl(ctx context.Context, camera *db.Camera, req *events.PtzCtrlRequest) error {
	logger.SInfo("requested to perform PTZ Control",
		zap.Any("request", req),
		zap.String("camera_id", req.CameraId))

	if err := s.requestRemoteControl(ctx, camera, req); err != nil {
		logger.SError("failed to perform PTZ Control", zap.Error(err))
		return err
	}

	logger.SInfo("PTZ Control success", zap.String("cameraId", req.CameraId))
	return nil
}

func (s *CommandService) requestRemoteControl(ctx context.Context, camera *db.Camera, req *events.PtzCtrlRequest) error {
	hasStopAfter := req.StopAfterSeconds != nil && *req.StopAfterSeconds > 0
	client := s.hikvisionClient.PtzCtrl(&hikvision.Credentials{
		Username: camera.Username,
		Password: camera.Password,
		Ip:       camera.Ip,
	})
	channelId := "1"
	continuousOptions := &hikvision.PtzCtrlContinousOptions{
		Pan:  req.Pan,
		Tilt: req.Tilt,
	}
	if hasStopAfter {
		if err := s.doContinousWithStop(ctx, client, &hikvision.PtzCtrlContinousWithResetRequest{
			ChannelId:  channelId,
			Options:    continuousOptions,
			ResetAfter: time.Second * time.Duration(*req.StopAfterSeconds),
		}); err != nil {
			return err
		}
	} else {
		if err := s.doRawContinuous(ctx, client, &hikvision.PtzCtrlRawContinousRequest{
			ChannelId: channelId,
			Options:   continuousOptions,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *CommandService) doRawContinuous(
	ctx context.Context,
	client hikvision.PtzApiClientInterface,
	options *hikvision.PtzCtrlRawContinousRequest) error {
	if err := client.RawContinuous(ctx, options); err != nil {
		return err
	}
	return nil
}

func (s *CommandService) doContinousWithStop(
	ctx context.Context,
	client hikvision.PtzApiClientInterface,
	options *hikvision.PtzCtrlContinousWithResetRequest) error {
	if err := client.ContinousWithReset(ctx, options); err != nil {
		return err
	}
	return nil
}
