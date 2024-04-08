package service

import (
	"context"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"

	"go.uber.org/zap"
)

type CommandService struct {
	hikvisionClient hikvision.Client
}

func NewCommandService(hikvisionClient hikvision.Client) *CommandService {
	return &CommandService{
		hikvisionClient: hikvisionClient,
	}
}

func (s *CommandService) Shutdown() {
}

func (s *CommandService) DeviceInfo(ctx context.Context, camera *db.Camera) (*hikvision.SystemDeviceInfoResponse, error) {
	logger.SDebug("requested retrieving device info",
		zap.Reflect("camera_id", camera.CameraId))
	info, err := s.hikvisionClient.System(&hikvision.Credentials{
		Ip:       camera.Ip,
		Username: camera.Username,
		Password: camera.Password,
	}).DeviceInfo(ctx)
	if err != nil {
		logger.SError("failed to retrieve device info",
			zap.Error(err))
		return nil, err
	}
	logger.SInfo("retrieved device info",
		zap.Reflect("info", info),
		zap.Reflect("camera_id", camera.CameraId))
	return info, nil
}
