package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
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
