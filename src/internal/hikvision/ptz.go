package hikvision

import (
	"context"
	"encoding/xml"
	"fmt"
	custhttp "github.com/CE-Thesis-2023/ltd/internal/http"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	"time"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type PtzApiClientInterface interface {
	Channels(ctx context.Context) (*PTZCtrlChannelsResponse, error)
	Capabilities(ctx context.Context, channelId string) (*PtzCtrlChannelCapabilities, error)
	RawContinuous(ctx context.Context, req *PtzCtrlRawContinousRequest) error
	ContinousWithReset(ctx context.Context, req *PtzCtrlContinousWithResetRequest) error
}

type ptzApiClient struct {
	restClient fastshot.ClientHttpMethods
	pool       *ants.Pool
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
	p := c.getBaseUrl()

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp PTZCtrlChannelsResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type PtzCtrlChannelCapabilities struct {
}

func (c *ptzApiClient) Capabilities(ctx context.Context, channelId string) (*PtzCtrlChannelCapabilities, error) {
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

	var parsedResp PtzCtrlChannelCapabilities
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
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
	p := fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId))

	body, err := xml.Marshal(req.Options)
	if err != nil {
		return err
	}

	resp, err := c.restClient.PUT(p).
		Context().Set(ctx).
		Body().AsString(string(body)).
		Send()
	if err != nil {
		return err
	}

	if err := handleError(&resp); err != nil {
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
	p := fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId))

	callbackChan := make(chan bool, 1)
	// do this first to make sure a goroutine is available, or else it blocks
	c.pool.Submit(func() {
		<-callbackChan
		<-time.After(req.ResetAfter)
		logger.SDebug("hikvision.ContinousWithReset: start reset now")

		resp, err := c.restClient.PUT(p).
			Context().Set(context.Background()).
			Body().AsString(ptzCtrlResetRequestBody).
			Retry().Set(2, time.Millisecond*2).
			Send()
		if err != nil {
			logger.SError("hikvision.ContinousWithReset: reset request sending error",
				zap.Error(err))
			return
		}

		if err := handleError(&resp); err != nil {
			logger.SError("hikvision.ContinousWithReset: reset request returned error",
				zap.Error(err))
			return
		}

		logger.SDebug("hikvision.ContinousWithReset: reset completed",
			zap.String("channelId", req.ChannelId),
			zap.Duration("after", req.ResetAfter))
	})

	forwardRequestBody, err := xml.Marshal(req.Options)
	if err != nil {
		return err
	}

	resp, err := c.restClient.PUT(p).
		Context().Set(ctx).
		Body().AsString(string(forwardRequestBody)).
		Send()
	if err != nil {
		return err
	}

	if err := handleError(&resp); err != nil {
		return err
	}

	callbackChan <- true
	logger.SDebug("hikvision.ContinousWithReset: first request completed")

	return nil
}
