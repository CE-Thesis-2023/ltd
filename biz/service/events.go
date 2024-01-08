package service

import (
	"context"
	"errors"
	"fmt"
	"labs/local-transcoder/helper/factory"
	"labs/local-transcoder/internal/cache"
	custdb "labs/local-transcoder/internal/db"
	custerror "labs/local-transcoder/internal/error"
	"labs/local-transcoder/internal/hikvision"
	"labs/local-transcoder/internal/logger"
	"labs/local-transcoder/internal/ome"
	"labs/local-transcoder/models/db"
	"labs/local-transcoder/models/events"
	"labs/local-transcoder/models/ms"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/dgraph-io/ristretto"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type CommandServiceInterface interface {
	PtzCtrl(ctx context.Context, req *events.PtzCtrlRequest) error
	DeviceInfo(ctx context.Context, req *events.CommandRetrieveDeviceInfo) error
	StreamChannels(ctx context.Context, req *events.CommandRetrieveStreamChannels) error
	StreamStatus(ctx context.Context, req *events.CommandGetStreamStatusRequest) error
	AddCamera(ctx context.Context, req *events.CommandAddCameraInfo) error
	StartStream(ctx context.Context, req *events.CommandStartStreamInfo) error
	EndStream(ctx context.Context, req *events.CommandEndStreamInfo) error
	StartFfmpegStream(ctx context.Context, req *events.CommandStartStreamInfo) error
	EndFfmpegStream(ctx context.Context, req *events.CommandEndStreamInfo) error
	Shutdown()
}
type CommandService struct {
	db              *custdb.LayeredDb
	cache           *ristretto.Cache
	omeClient       ome.OmeClientInterface
	hikvisionClient hikvision.Client
	pool            *ants.Pool

	streamManagementService StreamManagementServiceInterface
}

func NewCommandService() CommandServiceInterface {
	p, _ := ants.NewPool(10,
		ants.WithLogger(logger.NewZapToAntsLogger(logger.Logger())))
	return &CommandService{
		db:                      custdb.Layered(),
		cache:                   cache.Cache(),
		omeClient:               factory.Ome(),
		hikvisionClient:         factory.Hikvision(),
		pool:                    p,
		streamManagementService: GetStreamManagementService(),
	}
}

func (s *CommandService) Shutdown() {
	logger.SInfo("CommandService: shutdown requested")
	s.pool.Release()
}

func (s *CommandService) AddCamera(ctx context.Context, req *events.CommandAddCameraInfo) error {
	logger.SInfo("biz.AddCamera: request", zap.Any("request", req))
	camera, err := s.getCameraByName(ctx, req.Name)
	if err != nil {
		if !errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("AddCamera: getCameraByName", zap.Error(err))
			return err
		}
	}
	if camera != nil {
		logger.SError("AddCamera: camera already exists", zap.Any("camera", camera))
		return custerror.ErrorAlreadyExists
	}

	insertingCamera := db.Camera{
		Id:        req.CameraId,
		Name:      req.Name,
		Ip:        req.Ip,
		Port:      req.Port,
		Username:  req.Username,
		Password:  req.Password,
		DateAdded: time.Now().Format(time.RFC3339),
	}

	if err := s.saveCamera(ctx, insertingCamera); err != nil {
		logger.SError("AddCamera: saveCamera", zap.Error(err))
		return err
	}

	logger.SInfo("AddCamera: camera added successfully", zap.Any("camera", insertingCamera))
	return nil
}

func (s *CommandService) getCameraByName(ctx context.Context, name string) (*db.Camera, error) {
	val, found := s.cache.Get(fmt.Sprintf("camera-by_name-%s", name))
	if found {
		camera, yes := val.(db.Camera)
		if yes {
			logger.SDebug("getCameraByName: cache hit")
			return &camera, nil
		}
	}
	sqlExp := squirrel.Select("*").
		From("cameras").
		Where("name = ?", name)
	var camera db.Camera
	if err := s.db.Get(ctx, sqlExp, &camera); err != nil {
		logger.SError("getCameraByName: Get error", zap.Error(err))
		return nil, err
	}
	logger.SDebug("getCameraByName: db hit")
	s.pool.Submit(func() {
		set := s.cache.Set("camera-by_name-%s", camera, 100)
		if set {
			logger.SDebug("getCameraByName: cache set")
		} else {
			logger.SError("getCameraByName: cache set failed")
		}
	})
	return &camera, nil
}

func (s *CommandService) getCameraById(ctx context.Context, id string) (*db.Camera, error) {
	val, found := s.cache.Get(fmt.Sprintf("camera-by_id-%s", id))
	if found {
		camera, yes := val.(db.Camera)
		if yes {
			logger.SDebug("getCameraById: cache hit")
			return &camera, nil
		}
	}
	sqlExp := squirrel.Select("*").
		From("cameras").
		Where("id = ?", id)
	var camera db.Camera
	if err := s.db.Get(ctx, sqlExp, &camera); err != nil {
		logger.SError("getCameraById: Get error", zap.Error(err))
		return nil, err
	}
	logger.SDebug("getCameraById: db hit")
	s.pool.Submit(func() {
		set := s.cache.Set("camera-by_id-%s", camera, 100)
		if set {
			logger.SDebug("getCameraById: cache set")
		} else {
			logger.SError("getCameraById: cache set failed")
		}
	})
	return &camera, nil
}

func (s *CommandService) saveCamera(ctx context.Context, camera db.Camera) error {
	sqlExp := squirrel.Insert("cameras").
		Columns(camera.Fields()...).
		Values(camera.Values()...)
	if err := s.db.Insert(ctx, sqlExp); err != nil {
		logger.SError("saveCamera: db.Insert error", zap.Error(err))
		return err
	}
	s.pool.Submit(func() {
		successName := s.cache.Set(
			fmt.Sprintf("camera_by_name-%s", camera.Name),
			camera,
			100)
		successId := s.cache.Set(
			fmt.Sprintf("camera_by-id-%s", camera.Id),
			camera,
			101,
		)
		success := successName && successId
		if success {
			logger.SDebug("saveCamera: cache set")
		} else {
			logger.SError("saveCamera: cache set failed")
		}
	})
	logger.SDebug("saveCamera: saved to db")
	return nil
}

func (s *CommandService) DeviceInfo(ctx context.Context, req *events.CommandRetrieveDeviceInfo) error {
	panic("unimplemented")
}

func (s *CommandService) StartStream(ctx context.Context, req *events.CommandStartStreamInfo) error {
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("StartStream: camera not found",
				zap.String("id", req.CameraId),
				zap.Error(err))
			return err
		}
		logger.SError("StartStream: getCameraById error", zap.Error(err))
		return err
	}

	logger.SDebug("StartStream: camera", zap.Any("camera", camera))
	m := s.streamManagementService.MediaService()
	if err := m.RequestPullRtsp(ctx, camera, req); err != nil {
		logger.SError("StartStream: request stream error",
			zap.Error(err))
		return err
	}

	resp, err := m.RequestPushSrt(ctx, &ms.PushStreamingRequest{
		StreamName: req.CameraId,
	})
	if err != nil {
		logger.SError("RequestPushSrt: request push srt", zap.Error(err))
		return err
	}

	logger.SInfo("StartStream: sucess", zap.Any("pushStreaming", resp))
	return nil
}

func (s *CommandService) EndStream(ctx context.Context, req *events.CommandEndStreamInfo) error {
	return nil
}

func (s *CommandService) StartFfmpegStream(ctx context.Context, req *events.CommandStartStreamInfo) error {
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("StartFfmpegStream: camera not found",
				zap.String("id", req.CameraId),
				zap.Error(err))
			return err
		}
		logger.SError("StartFfmpegStream: getCameraById error", zap.Error(err))
		return err
	}

	logger.SDebug("StartFfmpegStream: camera", zap.Any("camera", camera))
	m := GetStreamManagementService().MediaService()
	err = m.RequestFFmpegRtspToSrt(ctx, camera, req)
	if err != nil {
		logger.SError("RequestFFmpegRtspToSrt: request push srt", zap.Error(err))
		return err
	}

	logger.SInfo("StartFfmpegStream: success", zap.Any("cameraId", req.CameraId))
	return nil
}

func (s *CommandService) EndFfmpegStream(ctx context.Context, req *events.CommandEndStreamInfo) error {
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("EndFfmpegStream: camera not found",
				zap.String("id", req.CameraId),
				zap.Error(err))
			return err
		}
		logger.SError("EndFfmpegStream: getCameraById error", zap.Error(err))
		return err
	}

	logger.SDebug("EndFfmpegStream: camera", zap.Any("camera", camera))
	m := GetStreamManagementService().MediaService()
	err = m.CancelFFmpegRtspToSrt(ctx, camera)
	if err != nil {
		logger.SError("CancelFFmpegRtspToSrt: cancel error", zap.Error(err))
		return nil
	}

	logger.SInfo("EndFfmpegStream: success", zap.Any("cameraId", req.CameraId))
	return nil
}

func (s *CommandService) StreamChannels(ctx context.Context, req *events.CommandRetrieveStreamChannels) error {
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("StreamChannels: camera not found",
				zap.String("id", req.CameraId),
				zap.Error(err))
			return err
		}
		logger.SError("StreamChannels: getCameraById error", zap.Error(err))
		return err
	}

	logger.SDebug("StreamChannels: camera", zap.Any("camera", camera))
	channelList, err := s.hikvisionClient.Streams(&hikvision.Credentials{
		Ip:       camera.Ip,
		Username: camera.Username,
		Password: camera.Password,
	}).Channels(ctx, &hikvision.StreamChannelsRequest{
		ChannelId: req.CameraId,
	})
	if err != nil {
		logger.SError("StreamChannels: Streams.Channels error", zap.Error(err))
		return err
	}

	logger.SInfo("StreamChannels: channelList", logger.Json("channels", channelList))
	return nil
}

func (s *CommandService) StreamStatus(ctx context.Context, req *events.CommandGetStreamStatusRequest) error {
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("StreamStatus: camera not found",
				zap.String("id", req.CameraId),
				zap.Error(err))
			return err
		}
		logger.SError("StreamStatus: getCameraById error", zap.Error(err))
		return err
	}

	logger.SDebug("StreamStatus: camera", zap.Any("camera", camera))
	statuses, err := s.hikvisionClient.Streams(&hikvision.Credentials{
		Ip:       camera.Ip,
		Username: camera.Username,
		Password: camera.Password,
	}).Status(ctx, &hikvision.StreamingStatusRequest{})
	if err != nil {
		logger.SError("StreamStatus: Streams.Status error", zap.Error(err))
		return err
	}

	logger.SInfo("StreamStatus: stream statuses", logger.Json("status", statuses))
	return nil
}
