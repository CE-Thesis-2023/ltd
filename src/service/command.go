package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math"
	"time"

	"github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/events"
	"github.com/CE-Thesis-2023/backend/src/models/ltdproxy"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/internal/opengate"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

type CommandService struct {
	hikvisionClient hikvision.Client
	MqttClient      *autopaho.ConnectionManager
	openGateClient  *opengate.OpenGateHTTPAPIClient
}

func NewCommandService(hikvisionClient hikvision.Client, mqttClient *autopaho.ConnectionManager, opengateClient *opengate.OpenGateHTTPAPIClient) *CommandService {
	return &CommandService{
		hikvisionClient: hikvisionClient,
		MqttClient:      mqttClient,
		openGateClient:  opengateClient,
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

func (s *CommandService) PtzCtrl(ctx context.Context, camera *db.Camera, req *events.PTZCtrlRequest) error {
	logger.SInfo("requested to perform PTZ Control",
		zap.Reflect("request", req),
		zap.String("camera_id", req.CameraId))

	if err := s.requestRemoteControl(ctx, camera, req); err != nil {
		logger.SError("failed to perform PTZ Control", zap.Error(err))
		return err
	}

	logger.SInfo("PTZ Control success", zap.String("cameraId", req.CameraId))
	return nil
}

func (s *CommandService) PTZCapabilties(ctx context.Context, camera *db.Camera) (*hikvision.PTZChannelCapabilities, error) {
	logger.SDebug("requested retrieving PTZ capabilities",
		zap.Reflect("camera_id", camera.CameraId))
	capabilities, err := s.hikvisionClient.PtzCtrl(&hikvision.Credentials{
		Ip:       camera.Ip,
		Username: camera.Username,
		Password: camera.Password,
	}).Capabilities(ctx, "1")
	if err != nil {
		logger.SError("failed to retrieve PTZ capabilities",
			zap.Error(err))
		return nil, err
	}
	logger.SInfo("retrieved PTZ capabilities",
		zap.Reflect("capabilities", capabilities),
		zap.Reflect("camera_id", camera.CameraId))
	return capabilities, nil
}

func (s *CommandService) requestRemoteControl(ctx context.Context, camera *db.Camera, req *events.PTZCtrlRequest) error {
	hasStopAfter := req.Duration > 0
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
			ResetAfter: time.Second * time.Duration(req.Duration),
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

type PTZStatusResponse struct {
	Moving bool `json:"moving"`
}

func (s *CommandService) PTZStatus(ctx context.Context, camera *db.Camera, req *hikvision.PtzCtrlStatusRequest) (*PTZStatusResponse, error) {
	ptzCtrl := s.hikvisionClient.PtzCtrl(&hikvision.Credentials{
		Username: camera.Username,
		Password: camera.Password,
		Ip:       camera.Ip,
	})

	status, err := ptzCtrl.Status(ctx, req)
	if err != nil {
		logger.SError("failed to retrieve PTZ status (1)", zap.Error(err))
		return nil, err
	}
	<-time.After(time.Millisecond * 500)
	updatedStatus, err := ptzCtrl.Status(ctx, req)
	if err != nil {
		logger.SError("failed to retrieve PTZ status (2)", zap.Error(err))
		return nil, err
	}

	azimuthDiff := 0.0
	elevationDiff := 0.0
	prev := status.AbsoluteHigh
	updated := updatedStatus.AbsoluteHigh

	azimuthDiff = math.Round(math.Abs(float64(updated.Azimuth-prev.Azimuth))*100) / 100
	elevationDiff = math.Round(math.Abs(float64(updated.Elevation-prev.Elevation))*100) / 100

	var resp PTZStatusResponse
	if azimuthDiff > 0.0 {
		resp.Moving = true
	}
	if elevationDiff > 0.0 {
		resp.Moving = true
	}
	return &resp, nil
}

func (s *CommandService) PTZRelative(ctx context.Context, camera *db.Camera, req *hikvision.PTZCtrlRelativeRequest) error {
	ptzCtrl := s.hikvisionClient.PtzCtrl(&hikvision.Credentials{
		Username: camera.Username,
		Password: camera.Password,
		Ip:       camera.Ip,
	})
	if err := ptzCtrl.Relative(ctx, req); err != nil {
		return err
	}
	return nil
}

func (s *CommandService) UploadEvent(ctx context.Context, req *ltdproxy.UploadEventRequest) error {
	if s.MqttClient == nil {
		return custerror.FormatInternalError("mqtt client is not initialized")
	}

	topic := req.Topic
	if len(topic) == 0 {
		return custerror.FormatInvalidArgument("missing topic")
	}
	eventId := req.
		Event.
		After.
		ID
	if len(eventId) == 0 {
		eventId = req.
			Event.
			Before.
			ID
	}
	if req.Snapshot == nil {
		jpegImg, err := s.openGateClient.EventsSnapshot(ctx, eventId, 480, 80)
		if err != nil {
			logger.SError("failed to retrieve snapshot", zap.Error(err))
			return err
		}
		base64Encoded := base64.StdEncoding.EncodeToString(jpegImg)
		req.Snapshot = &ltdproxy.EventSnapshot{
			Base64Image: base64Encoded,
		}
	}
	jsonPayload, err := json.Marshal(req)
	if err != nil {
		logger.SError("failed to marshal event", zap.Error(err))
		return err
	}
	if _, err := s.MqttClient.Publish(ctx, &paho.Publish{Topic: topic, Payload: jsonPayload}); err != nil {
		logger.SError("failed to publish event", zap.Error(err))
		return err
	}
	return nil
}
