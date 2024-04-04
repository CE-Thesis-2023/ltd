package service

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/CE-Thesis-2023/ltd/src/helper/factory"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	fastshot "github.com/opus-domini/fast-shot"

	"go.uber.org/zap"
)

type CommandService struct {
	hikvisionClient hikvision.Client

	backendHttpPrivateClient fastshot.ClientHttpMethods

	streamManagementService StreamManagementServiceInterface
}

func NewCommandService() *CommandService {
	configs := configs.Get().DeviceInfo
	backendClient := fastshot.NewClient(configs.CloudApiServer).
		Auth().BasicAuth(configs.Username, configs.Token).
		Config().SetCustomTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}).
		Build()
	return &CommandService{
		hikvisionClient:          factory.Hikvision(),
		streamManagementService:  GetStreamManagementService(),
		backendHttpPrivateClient: backendClient,
	}
}

func (s *CommandService) Shutdown() {
	return
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

func (s *CommandService) StartFfmpegStream(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	logger.SDebug("requested to start RTSP to SRT transcoding stream",
		zap.Reflect("camera_id", camera.CameraId),
		zap.Reflect("request", req))

	m := GetStreamManagementService().
		MediaService()
	err := m.RequestFFmpegRtspToSrt(ctx, camera, req)
	if err != nil {
		logger.SError("failed to start RTSP to SRT transcoding stream", zap.Error(err))
		return err
	}

	logger.SInfo("started transcoding stream",
		zap.Reflect("camera_id", req.CameraId))
	return nil
}

func (s *CommandService) EndFfmpegStream(ctx context.Context, camera *db.Camera, req *events.CommandEndStreamInfo) error {
	logger.SDebug("requested to end transcoding stream",
		zap.Reflect("camera", camera))

	m := GetStreamManagementService().
		MediaService()
	err := m.CancelFFmpegRtspToSrt(ctx, camera)
	if err != nil {
		logger.SError("failed to end RTSP to SRT transcoding stream",
			zap.Error(err))
		return nil
	}
	logger.SInfo("ended transcoding stream", zap.Reflect("cameraId", req.CameraId))
	return nil
}
