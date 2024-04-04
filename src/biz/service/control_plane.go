package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	fastshot "github.com/opus-domini/fast-shot"
)

type ControlPlaneService struct {
	client   fastshot.ClientHttpMethods
	basePath string
}

func NewControlPlaneService(configs *configs.DeviceInfoConfigs) *ControlPlaneService {
	builder := fastshot.NewClient(configs.CloudApiServer)
	clientConfigs := builder.Config()
	clientConfigs.SetTimeout(10 * time.Second)
	clientConfigs.SetFollowRedirects(true)
	clientConfigs.SetCustomTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	return &ControlPlaneService{client: builder.Build(), basePath: "/api/private"}
}

type RegistrationRequest struct {
	DeviceId string `json:"device_id"`
}

func (s *ControlPlaneService) RegisterDevice(ctx context.Context, req *RegistrationRequest) error {
	response, err := s.client.POST(s.basePath+"/registers").
		Context().Set(ctx).
		Body().AsJSON(req).
		Retry().Set(2, 5*time.Second).
		Send()
	if err != nil {
		return err
	}
	if response.Is2xxSuccessful() {
		return nil
	}
	if response.Is4xxClientError() {
		return custerror.ErrorAlreadyExists
	}
	return custerror.ErrorInternal
}

type GetAssignedDevicesRequest struct {
	DeviceId string `json:"device_id"`
}

type GetAssignedDevicesResponse struct {
	Cameras []db.Camera `json:"cameras"`
}

func (s *ControlPlaneService) GetAssignedDevices(ctx context.Context, req *GetAssignedDevicesRequest) (*GetAssignedDevicesResponse, error) {
	response, err := s.client.GET(s.basePath+"/transcoders/:id/cameras").
		Context().Set(ctx).
		Query().AddParam("id", req.DeviceId).
		Retry().Set(2, 5*time.Second).
		Send()
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, custerror.ErrorInternal
	}
	bodyBytes, err := io.ReadAll(response.RawBody())
	if err != nil {
		return nil, custerror.FormatInternalError("unable to read response body: %s", err)
	}
	var resp GetAssignedDevicesResponse
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
