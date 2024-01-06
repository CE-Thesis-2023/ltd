package hikvision

import (
	"context"
	"fmt"
	custhttp "labs/local-transcoder/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
)

type SystemApiInterface interface {
	Capabilities(ctx context.Context) (*SystemCapabilitiesResponse, error)
	DeviceInfo(ctx context.Context) (*SystemDeviceInfoResponse, error)
	Hardware(ctx context.Context) (*SystemHardwareResponse, error)
}

type systemApiClient struct {
	restClient fastshot.ClientHttpMethods
	pool       *ants.Pool
}

func (c *systemApiClient) getBaseUrl() string {
	return "/System"
}

type SystemCapabilitiesResponse struct {
}

func (c *systemApiClient) Capabilities(ctx context.Context) (*SystemCapabilitiesResponse, error) {
	p := fmt.Sprintf("%s/capabilities", c.getBaseUrl())

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemCapabilitiesResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type SystemDeviceInfoResponse struct {
}

func (c *systemApiClient) DeviceInfo(ctx context.Context) (*SystemDeviceInfoResponse, error) {
	p := fmt.Sprintf("%s/deviceinfo", c.getBaseUrl())

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemDeviceInfoResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type SystemHardwareResponse struct {
}

func (c *systemApiClient) Hardware(ctx context.Context) (*SystemHardwareResponse, error) {
	p := fmt.Sprintf("%s/Hardware", c.getBaseUrl())
	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemHardwareResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}
