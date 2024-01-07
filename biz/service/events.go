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
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/dgraph-io/ristretto"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type CommandServiceInterface interface {
	PtzCtrl(ctx context.Context, req *events.PtzCtrlRequest) error
	DeviceInfo(ctx context.Context, req *events.CommandRetrieveDeviceInfo) error
	AddCamera(ctx context.Context, req *events.CommandAddCameraInfo) error
	StartStream(ctx context.Context, req *events.CommandStartStreamInfo) error
	EndStream(ctx context.Context, req *events.CommandEndStreamInfo) error
}
type CommandService struct {
	db              *custdb.LayeredDb
	cache           *ristretto.Cache
	omeClient       ome.OmeClientInterface
	hikvisionClient hikvision.Client
	pool            *ants.Pool
}

func NewCommandService() CommandServiceInterface {
	p, _ := ants.NewPool(10,
		ants.WithLogger(logger.NewZapToAntsLogger(logger.Logger())))
	return &CommandService{
		db:              custdb.Layered(),
		cache:           cache.Cache(),
		omeClient:       factory.Ome(),
		hikvisionClient: factory.Hikvision(),
		pool:            p,
	}
}

func (s *CommandService) AddCamera(ctx context.Context, req *events.CommandAddCameraInfo) error {
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
	if err := s.db.Select(ctx, sqlExp, &camera); err != nil {
		logger.SError("getCameraByName: select error", zap.Error(err))
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
	if err := s.db.Select(ctx, sqlExp, &camera); err != nil {
		logger.SError("getCameraById: select error", zap.Error(err))
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
	if err := s.requestStream(ctx, camera, req); err != nil {
		logger.SError("StartStream: request stream error",
			zap.Error(err))
		return err
	}

	logger.SInfo("StartStream: sucess")
	return nil
}

func (s *CommandService) requestStream(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	omeListResp, err := s.omeClient.Streams().List(ctx)
	if err != nil {
		logger.SDebug("requestStream: list streams error", zap.Error(err))
		return err
	}

	for _, stream := range omeListResp.Names {
		if strings.EqualFold(stream, req.CameraId) {
			logger.SError("requestStream: stream already started", zap.String("name", req.CameraId))
			return custerror.ErrorAlreadyExists
		}
	}

	logger.SDebug("requestStream: stream not started, attempting to ask media server to enable stream")

	if err := s.startPullRtspStream(ctx, camera, req); err != nil {
		logger.SError("requestStream: startPullRtspStream error", zap.Error(err))
		return err
	}

	logger.SInfo("requestStream: success")
	return nil
}

func (s *CommandService) startPullRtspStream(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	if err := s.omeClient.Streams().CreatePull(ctx, &ome.StreamCreationRequest{
		Name: req.CameraId,
		URLs: []string{
			s.buildRtspStreamUrl(camera, req),
		},
		Properties: ome.StreamProperties{
			Persistent:            true, // dont delete stream if no viewer or no input
			IgnoreRtcpSRTimestamp: false,
		},
	}); err != nil {
		logger.SDebug("startPullRtspStream: CreatePull error", zap.Error(err))
		return err
	}
	logger.SDebug("startPullRtspStream: success")
	return nil
}

func (s *CommandService) buildRtspStreamUrl(camera *db.Camera, req *events.CommandStartStreamInfo) string {
	u := &url.URL{}
	u.Scheme = "rtsp"
	u.Host = camera.Ip
	if camera.Port != 0 {
		u.Host = fmt.Sprintf("%s:%d", camera.Ip, camera.Port)
	}
	u = u.JoinPath("/ISAPI", "/Streaming", "channels", req.ChannelId)
	u.User = url.UserPassword(camera.Username, camera.Password)
	url := u.String()
	logger.SDebug("buildRtspStreamUrl: stream url", zap.String("url", url))
	return url
}

func (s *CommandService) EndStream(ctx context.Context, req *events.CommandEndStreamInfo) error {
	return nil
}
