package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	db "github.com/CE-Thesis-2023/backend/src/models/db"
	"github.com/CE-Thesis-2023/backend/src/models/web"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"
)

type ControlPlaneService struct {
	baseUrl           *url.URL
	httpClient        *http.Client
	basicAuthUser     string
	basicAuthPassword string
}

func NewControlPlaneService(configs *configs.DeviceInfoConfigs) *ControlPlaneService {
	baseUrl, err := url.Parse(configs.CloudApiServer + "/private")
	if err != nil {
		logger.SFatal("unable to parse base url", zap.Error(err))
	}
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	return &ControlPlaneService{
		baseUrl:           baseUrl,
		httpClient:        &client,
		basicAuthUser:     configs.Username,
		basicAuthPassword: configs.Password}
}

type RegistrationRequest struct {
	DeviceId string `json:"deviceId"`
}

func (s *ControlPlaneService) RegisterDevice(ctx context.Context, req *RegistrationRequest) error {
	logger.SDebug("registering device",
		zap.String("device_id", req.DeviceId),
		zap.Reflect("request", req))

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return custerror.FormatInternalError("unable to marshal request: %s", err)
	}
	body := bytes.NewReader(bodyBytes)
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		s.baseUrl.String()+"/registers",
		body)
	if err != nil {
		return custerror.FormatInternalError("unable to create http request: %s", err)
	}
	httpRequest.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient.Do(httpRequest)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 200:
		return nil
	case 400:
		return custerror.ErrorInvalidArgument
	case 409:
		return custerror.ErrorAlreadyExists
	default:
		return custerror.ErrorInternal
	}
}

type GetAssignedDevicesRequest struct {
	DeviceId string `json:"deviceId"`
}

type GetAssignedDevicesResponse struct {
	Cameras []db.Camera `json:"cameras"`
}

func (s *ControlPlaneService) GetAssignedDevices(ctx context.Context, req *GetAssignedDevicesRequest) (*GetAssignedDevicesResponse, error) {
	path := s.baseUrl.JoinPath(fmt.Sprintf("/transcoders/%s/cameras", req.DeviceId))
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp GetAssignedDevicesResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) GetOpenGateIntegrationConfigurations(ctx context.Context, req *web.GetOpenGateIntegrationByIdRequest) (*web.GetOpenGateCameraSettingsResponse, error) {
	path := s.baseUrl.JoinPath("/opengate", req.OpenGateId)
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp web.GetOpenGateCameraSettingsResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) GetOpenGateCameraSettings(ctx context.Context, req *web.GetOpenGateCameraSettingsRequest) (*web.GetOpenGateCameraSettingsResponse, error) {
	path := s.baseUrl.JoinPath("/opengate/cameras")
	q := path.Query()
	aggr := strings.Join(req.CameraId, ",")
	q.Add("camera_id", aggr)
	path.RawQuery = q.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp web.GetOpenGateCameraSettingsResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) GetOpenGateMqttConfigurations(ctx context.Context, req *web.GetOpenGateMqttSettingsRequest) (*web.GetOpenGateMqttSettingsResponse, error) {
	path := s.baseUrl.JoinPath("/opengate", req.ConfigurationId, "mqtt")
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp web.GetOpenGateMqttSettingsResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) GetOpenGateConfigurations(ctx context.Context, req *web.GetTranscoderOpenGateConfigurationRequest) (*web.GetTranscoderOpenGateConfigurationResponse, error) {
	path := s.baseUrl.JoinPath("/opengate", "configurations", req.TranscoderId)
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp web.GetTranscoderOpenGateConfigurationResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) GetCameraStreamSettings(ctx context.Context, req *web.GetStreamConfigurationsRequest) (*web.GetStreamConfigurationsResponse, error) {
	path := s.baseUrl.JoinPath("/transcoders/streams")
	q := path.Query()
	aggr := strings.Join(req.CameraId, ",")
	q.Add("camera_id", aggr)
	path.RawQuery = q.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp web.GetStreamConfigurationsResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) UpdateTranscoderStatus(ctx context.Context, transcoderId string, cameraId string, status bool) error {
	path := s.baseUrl.JoinPath("/transcoders/status")
	q := path.Query()
	q.Add("transcoder_id", transcoderId)
	q.Add("camera_id", cameraId)
	q.Add("transcoder_status", fmt.Sprintf("%t", status))
	path.RawQuery = q.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		path.String(),
		nil)
	if err != nil {
		return err
	}

	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 200, 201, 202:
		return nil
	case 400:
		return custerror.ErrorInvalidArgument
	case 404:
		return custerror.ErrorNotFound
	default:
		return custerror.ErrorInternal
	}
}

func (s *ControlPlaneService) GetMQTTEndpoints(ctx context.Context, req *web.GetMQTTEventEndpointRequest) (*web.GetMQTTEventEndpointResponse, error) {
	path := s.baseUrl.JoinPath("/transcoders/mqtt")
	q := path.Query()
	q.Add("transcoder_id", req.TranscoderId)
	path.RawQuery = q.Encode()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		path.String(),
		nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(s.basicAuthUser, s.basicAuthPassword)
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	switch response.StatusCode {
	case 200:
		defer response.
			Body.
			Close()
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, custerror.FormatInternalError("unable to read response body: %s", err)
		}
		var resp web.GetMQTTEventEndpointResponse
		if err := json.Unmarshal(bodyBytes, &resp); err != nil {
			return nil, err
		}
		return &resp, nil
	case 400:
		return nil, custerror.ErrorInvalidArgument
	case 404:
		return nil, custerror.ErrorNotFound
	default:
		return nil, custerror.ErrorInternal
	}
}
