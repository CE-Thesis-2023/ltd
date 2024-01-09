package hikvision

import (
	"context"
	"fmt"
	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
)

type SmartApiInterface interface {
	Capabilities(ctx context.Context) (*SmartCapabilitiesResponse, error)
}

type smartApiClient struct {
	restClient  fastshot.ClientHttpMethods
	pool        *ants.Pool
}

func (c *smartApiClient) getBaseUrl() string {
	return "/Smart"
}

type SmartCapabilitiesResponse struct {
}

func (c *smartApiClient) Capabilities(ctx context.Context) (*SmartCapabilitiesResponse, error) {
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

	var parsedResp SmartCapabilitiesResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}
