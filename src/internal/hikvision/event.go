package hikvision

import (
	"context"
	"fmt"
	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
)

type EventApiInterface interface {
	Capabilities(ctx context.Context) (*EventCapabilitiesResponse, error)
	CapabilitiesChannel(ctx context.Context, channelId string) (*EventCapabilitiesOfChannelResponse, error)
}

type eventApiClient struct {
	restClient fastshot.ClientHttpMethods
	pool       *ants.Pool
}

func (c *eventApiClient) getBaseUrl() string {
	return "/Event"
}

func (c *eventApiClient) getUrlWithChannel(id string) string {
	return fmt.Sprintf("%s/%s", c.getBaseUrl(), id)
}

type EventCapabilitiesResponse struct {
}

func (c *eventApiClient) Capabilities(ctx context.Context) (*EventCapabilitiesResponse, error) {
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

	var parsedResp EventCapabilitiesResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type EventCapabilitiesOfChannelResponse struct {
}

func (c *eventApiClient) CapabilitiesChannel(ctx context.Context, channelId string) (*EventCapabilitiesOfChannelResponse, error) {
	p := fmt.Sprintf("%s/capabilities", c.getUrlWithChannel(channelId))

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp EventCapabilitiesOfChannelResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}
