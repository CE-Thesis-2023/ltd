package hikvision

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
)

type SmartApiInterface interface {
	Capabilities(ctx context.Context) (*SmartCapabilitiesResponse, error)
}

type smartApiClient struct {
	httpClient *http.Client
	username   string
	password   string
	ip         string
}

func (c *smartApiClient) getBaseUrl() string {
	return c.ip + "/Smart"
}

type SmartCapabilitiesResponse struct {
}

func (c *smartApiClient) Capabilities(ctx context.Context) (*SmartCapabilitiesResponse, error) {
	p, _ := url.Parse(fmt.Sprintf("%s/capabilities", c.getBaseUrl()))

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

	var parsedResp SmartCapabilitiesResponse
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}
