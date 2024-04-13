package hikvision

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
)

type EventApiInterface interface {
	Capabilities(ctx context.Context) (*EventCapabilitiesResponse, error)
	CapabilitiesChannel(ctx context.Context, channelId string) (*EventCapabilitiesOfChannelResponse, error)
}

type eventApiClient struct {
	httpClient *http.Client
	username   string
	password   string
	ip         string
}

func (c *eventApiClient) getBaseUrl() string {
	return c.ip + "/Event"
}

func (c *eventApiClient) getUrlWithChannel(id string) string {
	return fmt.Sprintf("%s/%s", c.getBaseUrl(), id)
}

type EventCapabilitiesResponse struct {
}

func (c *eventApiClient) Capabilities(ctx context.Context) (*EventCapabilitiesResponse, error) {
	p, _ := url.Parse(fmt.Sprintf("%s/capabilities", c.getBaseUrl()))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password))
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if err := handleError(response); err != nil {
		return nil, err
	}

	var parsedResp EventCapabilitiesResponse
	if err := custhttp.XMLResponse(response, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type EventCapabilitiesOfChannelResponse struct {
}

func (c *eventApiClient) CapabilitiesChannel(ctx context.Context, channelId string) (*EventCapabilitiesOfChannelResponse, error) {
	p, _ := url.Parse(fmt.Sprintf("%s/capabilities", c.getUrlWithChannel(channelId)))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if err := handleError(resp); err != nil {
		return nil, err
	}

	var parsedResp EventCapabilitiesOfChannelResponse
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}
