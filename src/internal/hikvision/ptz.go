package hikvision

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"

	"go.uber.org/zap"
)

type PtzApiClientInterface interface {
	Channels(ctx context.Context) (*PTZCtrlChannelsResponse, error)
	Capabilities(ctx context.Context, channelId string) (*PtzCtrlChannelCapabilities, error)
	RawContinuous(ctx context.Context, req *PtzCtrlRawContinousRequest) error
	ContinousWithReset(ctx context.Context, req *PtzCtrlContinousWithResetRequest) error
}

type ptzApiClient struct {
	httpClient *http.Client
	username   string
	password   string
}

func (c *ptzApiClient) getBaseUrl() string {
	return "/PTZCtrl/channels"
}

func (c *ptzApiClient) getUrlWithChannel(id string) string {
	return fmt.Sprintf("%s/%s", c.getBaseUrl(), id)
}

type PTZCtrlChannelsResponse struct {
}

func (c *ptzApiClient) Channels(ctx context.Context) (*PTZCtrlChannelsResponse, error) {
	p, _ := url.Parse(c.getBaseUrl())

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password),
	)
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

	var parsedResp PTZCtrlChannelsResponse
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type PtzCtrlChannelCapabilities struct {
}

func (c *ptzApiClient) Capabilities(ctx context.Context, channelId string) (*PtzCtrlChannelCapabilities, error) {
	p, _ := url.Parse(fmt.Sprintf("%s/capabilities", c.getUrlWithChannel(channelId)))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password),
	)
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

	var parsedResp PtzCtrlChannelCapabilities
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type PtzCtrlRawContinousRequest struct {
	ChannelId string
	Options   *PtzCtrlContinousOptions
}

type PtzCtrlContinousOptions struct {
	XMLName xml.Name `xml:"PTZData"`
	Pan     int      `xml:"pan"`  // + is right, - is left
	Tilt    int      `xml:"tilt"` // + is up, - is down
}

func (c *ptzApiClient) RawContinuous(ctx context.Context, req *PtzCtrlRawContinousRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId)))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(req.Options),
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}

	if err := handleError(resp); err != nil {
		return err
	}

	return nil
}

type PtzCtrlContinousWithResetRequest struct {
	ChannelId  string
	Options    *PtzCtrlContinousOptions
	ResetAfter time.Duration
}

var ptzCtrlResetRequestBody string = "<PTZData><pan>0</pan><tilt>0</tilt></PTZData>"

func (c *ptzApiClient) ContinousWithReset(ctx context.Context, req *PtzCtrlContinousWithResetRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId)))

	resetRequest, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(ptzCtrlResetRequestBody),
	)
	if err != nil {
		return err
	}

	continousRequest, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(req.Options),
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(continousRequest)
	if err != nil {
		return err
	}

	if err := handleError(resp); err != nil {
		return err
	}

	go func() {
		<-time.After(req.ResetAfter)
		logger.SDebug("PTZ Control continuous request start")

		resp, err := c.httpClient.Do(resetRequest)
		if err != nil {
			logger.SError("failed to send PTZ Control continuous reset request",
				zap.Error(err))
			return
		}

		if err := handleError(resp); err != nil {
			logger.SError("failed to send PTZ Control continuous reset request",
				zap.Error(err))
			return
		}

		logger.SDebug("PTZ Control continuous request reset completed",
			zap.String("channelId", req.ChannelId),
			zap.Duration("after", req.ResetAfter))
	}()

	return nil
}
