package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/CE-Thesis-2023/ltd/src/helper/factory"
	"github.com/CE-Thesis-2023/ltd/src/internal/cache"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custdb "github.com/CE-Thesis-2023/ltd/src/internal/db"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/CE-Thesis-2023/ltd/src/models/rest"
	"github.com/bytedance/sonic"
	fastshot "github.com/opus-domini/fast-shot"

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
	StartFfmpegStream(ctx context.Context, req *events.CommandStartStreamInfo) error
	EndFfmpegStream(ctx context.Context, req *events.CommandEndStreamInfo) error
	DebugListStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error)
	DeleteCamera(ctx context.Context, req *events.CommandDeleteCameraRequest) error
	UpdateCameraList(ctx context.Context) error
	RegisterDevice(ctx context.Context) error
	StartAllEnabledStreams(ctx context.Context) error
	Shutdown()
}

type CommandService struct {
	db              *custdb.LayeredDb
	cache           *ristretto.Cache
	hikvisionClient hikvision.Client
	pool            *ants.Pool

	backendHttpPrivateClient fastshot.ClientHttpMethods

	streamManagementService StreamManagementServiceInterface
}

func NewCommandService() CommandServiceInterface {
	configs := configs.Get().DeviceInfo
	p, _ := ants.NewPool(10,
		ants.WithLogger(logger.NewZapToAntsLogger(logger.Logger())))
	backendClient := fastshot.NewClient(configs.CloudApiServer).
		Auth().BasicAuth(configs.Username, configs.Token).
		Config().SetCustomTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}).
		Build()
	return &CommandService{
		db:                       custdb.Layered(),
		cache:                    cache.Cache(),
		hikvisionClient:          factory.Hikvision(),
		pool:                     p,
		streamManagementService:  GetStreamManagementService(),
		backendHttpPrivateClient: backendClient,
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
		CameraId: req.CameraId,
		Name:     req.Name,
		Ip:       req.Ip,
		Port:     req.Port,
		Username: req.Username,
		Password: req.Password,
		Started:  false,
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
	sqlExp := squirrel.Select("*").
		From("cameras").
		Where("camera_id = ?", id)
	var camera db.Camera
	if err := s.db.Get(ctx, sqlExp, &camera); err != nil {
		logger.SError("getCameraById: Get error", zap.Error(err))
		return nil, err
	}
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
	logger.SDebug("saveCamera: saved to db")
	return nil
}

func (s *CommandService) DeviceInfo(ctx context.Context, req *events.CommandRetrieveDeviceInfo) error {
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("DeviceInfo: camera not found",
				zap.String("id", req.CameraId),
				zap.Error(err))
			return err
		}
		logger.SError("DeviceInfo: getCameraById error", zap.Error(err))
		return err
	}

	logger.SDebug("DeviceInfo: camera", zap.Any("camera", camera))
	info, err := s.hikvisionClient.System(&hikvision.Credentials{
		Ip:       camera.Ip,
		Username: camera.Username,
		Password: camera.Password,
	}).DeviceInfo(ctx)
	if err != nil {
		logger.SError("StreamStatus: System.DeviceInfo error", zap.Error(err))
		return err
	}

	logger.SInfo("StreamStatus: device info", logger.Json("info", info))
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
		ChannelId: "",
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

func (s *CommandService) DebugListStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error) {
	logger.SInfo("DebugListStreams: request")
	streams, err := s.streamManagementService.
		MediaService().
		ListOngoingStreams(ctx)
	if err != nil {
		logger.SDebug("DebugListStreams: sms.ListOngoingStreams error", zap.Error(err))
		return nil, err
	}
	return streams, nil
}

func (s *CommandService) DeleteCamera(ctx context.Context, req *events.CommandDeleteCameraRequest) error {
	logger.SInfo("DeleteCamera: request")
	camera, err := s.getCameraById(ctx, req.CameraId)
	if err != nil {
		logger.SDebug("DeleteCamera: getCameraById error", zap.Error(err))
		return err
	}

	if err := s.deleteCamera(ctx, camera.CameraId); err != nil {
		logger.SError("DeleteCamera: deleteCamera error", zap.Error(err))
		return err
	}

	if err := s.streamManagementService.MediaService().CancelFFmpegRtspToSrt(ctx, camera); err != nil {
		logger.SError("DeleteCamera: stop stream error", zap.Error(err))
		return err
	}

	logger.SInfo("DeleteCamera: camera deleted", zap.String("id", req.CameraId))
	return nil
}

func (s *CommandService) deleteCamera(ctx context.Context, id string) error {
	q := squirrel.Delete("cameras").Where("camera_id = ?", id)
	return s.db.Delete(ctx, q)
}

func (s *CommandService) RegisterDevice(ctx context.Context) error {
	logger.SDebug("RegisterDevice: starting")
	deviceId := configs.Get().DeviceInfo.DeviceId

	req := map[string]interface{}{
		"deviceId": deviceId,
	}

	resp, err := s.backendHttpPrivateClient.POST("/registers").
		Body().AsJSON(req).
		Context().Set(ctx).
		Send()
	if err != nil {
		logger.SError("client.POST /registers: error", zap.Error(err))
		return err
	}

	if resp.Is2xxSuccessful() {
		logger.SInfo("RegisterDevice: success 200 OK")
		return nil
	}

	if resp.StatusCode() == http.StatusConflict {
		return custerror.ErrorAlreadyExists
	}

	logger.SError("RegisterDevice: received 4xx or 5xx",
		zap.String("statusText", resp.StatusText()))
	return custerror.ErrorInternal
}

func (s *CommandService) UpdateCameraList(ctx context.Context) error {
	logger.SDebug("UpdateCameraList: starting")
	deviceId := configs.Get().DeviceInfo.DeviceId

	resp, err := s.backendHttpPrivateClient.GET(fmt.Sprintf("/transcoders/%s/cameras", deviceId)).
		Context().Set(ctx).
		Send()
	if err != nil {
		logger.SError("client.GET /transcoders/%s/cameras", zap.Error(err))
		return err
	}

	receivedCameras := []db.Camera{}
	if resp.Is2xxSuccessful() {
		respMap := map[string]interface{}{}
		if err := custhttp.JSONResponse(&resp, &respMap); err != nil {
			logger.SError("UpdateCameraList: JSONResponse error", zap.Error(err))
			return err
		}

		cameraList := respMap["cameras"]
		respContent, _ := sonic.Marshal(&cameraList)
		if err := sonic.Unmarshal(respContent, &receivedCameras); err != nil {
			logger.SError("mapstructure.Decode: error", zap.Error(err))
			return err
		}
	}

	if resp.Is5xxServerError() {
		logger.SInfo("UpdateCameraList: status not 200", zap.String("statusText", resp.StatusText()))
		return custerror.ErrorInternal
	}

	currentCameras, err := s.getCameras(ctx)
	if err != nil {
		logger.SError("getCameras: error", zap.Error(err))
		return err
	}

	if err := s.addUpdateOrDeleteCameras(ctx, currentCameras, receivedCameras); err != nil {
		logger.SError("addUpdateOrDeleteCameras: error", zap.Error(err))
		return err
	}

	logger.SInfo("UpdateCameraList: migrated camera list success")
	return nil
}

func (s *CommandService) getCameras(ctx context.Context) ([]db.Camera, error) {
	q := squirrel.Select("*").From("cameras")
	list := []db.Camera{}
	if err := s.db.Select(ctx, q, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *CommandService) updateCamera(ctx context.Context, camera *db.Camera) error {
	valueMap := map[string]interface{}{}
	fields := camera.Fields()
	values := camera.Values()
	for i := 0; i < len(fields); i += 1 {
		valueMap[fields[i]] = values[i]
	}

	q := squirrel.Update("cameras").Where("camera_id = ?", camera.CameraId).SetMap(valueMap)
	sql, args, _ := q.ToSql()
	logger.SDebug("updateCamera: SQL query",
		zap.String("query", sql),
		zap.Any("args", args))
	if err := s.db.Update(ctx, q); err != nil {
		return err
	}
	return nil
}

func (s *CommandService) addUpdateOrDeleteCameras(ctx context.Context, currentList []db.Camera, updatedList []db.Camera) error {
	updatedMap := map[string]db.Camera{}
	for _, c := range updatedList {
		updatedMap[c.CameraId] = c
	}
	for _, current := range currentList {
		id := current.CameraId
		logger.SDebug("addUpdateOrDeleteCameras: updating camera", zap.String("id", id))
		updated, found := updatedMap[id]
		if !found {
			if err := s.deleteCamera(ctx, current.CameraId); err != nil {
				logger.SError("deleteCamera: error", zap.Error(err))
				return err
			}
			logger.SDebug("addUpdateOrDeleteCameras: deleted camera", zap.String("id", id))
		}
		if err := s.updateCamera(ctx, &updated); err != nil {
			logger.SError("updateCamera: error", zap.Error(err))
			return err
		}
		logger.SDebug("addUpdateOrDeleteCameras: updated camera", zap.String("id", id))
	}

	currentMap := map[string]db.Camera{}
	for _, updated := range updatedList {
		id := updated.CameraId
		_, found := currentMap[id]
		if !found {
			if err := s.saveCamera(ctx, updated); err != nil {
				logger.SError("saveCamera: error", zap.Error(err))
				return err
			}
			logger.SDebug("addUpdateOrDeleteCameras: added camera", zap.String("id", id))
		}
	}

	return nil
}

func (s *CommandService) StartAllEnabledStreams(ctx context.Context) error {
	logger.SInfo("StartAllEnabledStreams: started")

	streams, err := s.getEnabledStreams(ctx)
	if err != nil {
		if errors.Is(err, custerror.ErrorNotFound); err != nil {
			logger.SDebug("StartAllEnabledStreams: no enabled or started streams found")
			return nil
		}
		logger.SError("StartAllEnabledStreams: error", zap.Error(err))
		return err
	}

	for _, stream := range streams {
		if stream.Started {
			if err := s.streamManagementService.MediaService().RequestFFmpegRtspToSrt(ctx, &stream, &events.CommandStartStreamInfo{
				CameraId:  stream.CameraId,
				ChannelId: "1",
			}); err != nil {
				logger.SError("StartAllEnabledStreams: RequestFfmpegRtspToSrt failed",
					zap.Error(err),
					zap.String("cameraId", stream.CameraId))
				return err
			}
			logger.SInfo("StartAllEnabledStreams: stream started",
				zap.String("id", stream.CameraId),
				zap.String("ip", stream.Ip))
		}
	}

	logger.SInfo("StartAllEnabledStreams: completed")
	return nil
}

func (s *CommandService) getEnabledStreams(ctx context.Context) ([]db.Camera, error) {
	logger.SDebug("getEnabledStreams: started")

	q := squirrel.Select("*").From("cameras").Where("started = ?", true)

	var enabledStreams []db.Camera
	if err := s.db.Select(ctx, q, &enabledStreams); err != nil {
		logger.SError("getEnabledStreams: Select error", zap.Error(err))
		return nil, err
	}

	return enabledStreams, nil
}
