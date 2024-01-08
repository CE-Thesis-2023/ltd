package ome

import (
	"context"
	"fmt"
	"labs/local-transcoder/internal/configs"
	custerror "labs/local-transcoder/internal/error"
	custhttp "labs/local-transcoder/internal/http"
	"time"

	fastshot "github.com/opus-domini/fast-shot"
)

type OmePushApiInterface interface {
	Start(ctx context.Context, req *StartPushStreamingRequest) (*StartPushStreamingResponse, error)
	Stop(ctx context.Context, req *PushIdRequest) error
	List(ctx context.Context) (*[]PushStreamingInfo, error)
}

type omePushApiClient struct {
	localConfigs *configs.OvenMediaEngineConfigs
	restClient   fastshot.ClientHttpMethods
}

var _ OmePushApiInterface = (*omePushApiClient)(nil)

type OriginStreamInfo struct {
	Name         string   `json:"name"`
	VariantNames []string `json:"variantNames"`
}

type Protocol string

const (
	ProtocolSRT    Protocol = "srt"
	ProtocolRTMP   Protocol = "rtmp"
	ProtocolMPEGTS Protocol = "mpegts"
)

type StartPushStreamingRequest struct {
	ID        string           `json:"id"`
	Stream    OriginStreamInfo `json:"stream"`
	Protocol  Protocol         `json:"protocol"`
	URL       string           `json:"url"`
	StreamKey *string          `json:"streamKey,omitempty"`
}

type StreamInfo struct {
	Name         string   `json:"name"`
	TrackIds     []string `json:"trackIds"`
	VariantNames []string `json:"variantNames"`
}

type StartPushStreamingResponse struct {
	ID             string     `json:"id"`
	State          string     `json:"state"`
	Vhost          string     `json:"vhost"`
	App            string     `json:"app"`
	Stream         StreamInfo `json:"stream"`
	Protocol       string     `json:"protocol"`
	URL            string     `json:"url"`
	StreamKey      string     `json:"streamKey"`
	SentBytes      int        `json:"sentBytes"`
	SentTime       int        `json:"sentTime"`
	Sequence       int        `json:"sequence"`
	TotalSentBytes int        `json:"totalsentBytes"`
	TotalSentTime  int        `json:"totalsentTime"`
	CreatedTime    string     `json:"createdTime"`
	StartTime      string     `json:"startTime"`
	FinishTime     string     `json:"finishTime"`
}

type PushApiCommonResponse[T interface{}] struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Response   *T     `json:"response,omitempty"`
}

func (c *omePushApiClient) getBaseUrl() string {
	return fmt.Sprintf(
		"/v1/vhosts/%s/apps/%s",
		c.localConfigs.VirtualHostName,
		c.localConfigs.ApplicationName)
}

func (c *omePushApiClient) handleError(ctx context.Context, resp *fastshot.Response) error {
	switch resp.StatusCode() {
	case 400:
		var parsedResp PushApiCommonResponse[interface{}]
		if err := custhttp.JSONResponse(resp, &parsedResp); err != nil {
			return err
		}
		return custerror.FormatInvalidArgument(parsedResp.Message)
	case 401:
		return custerror.ErrorPermissionDenied
	case 404:
		return custerror.ErrorNotFound
	case 409:
		var parsedResp PushApiCommonResponse[interface{}]
		if err := custhttp.JSONResponse(resp, &parsedResp); err != nil {
			return err
		}
		return custerror.FormatAlreadyExists(parsedResp.Message)
	case 500:
		return custerror.ErrorInternal
	}

	return nil
}

func (c *omePushApiClient) Start(ctx context.Context, req *StartPushStreamingRequest) (*StartPushStreamingResponse, error) {
	p := fmt.Sprintf("%s:startPush", c.getBaseUrl())
	resp, err := c.restClient.POST(p).
		Body().AsJSON(req).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := c.handleError(ctx, &resp); err != nil {
		return nil, err
	}

	var parsedResp PushApiCommonResponse[StartPushStreamingResponse]
	if err := custhttp.JSONResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return parsedResp.Response, nil
}

type PushIdRequest struct {
	Id string `json:"id"`
}

func (c *omePushApiClient) Stop(ctx context.Context, req *PushIdRequest) error {
	p := fmt.Sprintf("%s:stopPush", c.getBaseUrl())
	resp, err := c.restClient.POST(p).
		Body().AsJSON(req).
		Context().Set(ctx).
		Send()
	if err != nil {
		return err
	}

	if err := c.handleError(ctx, &resp); err != nil {
		return err
	}

	return nil
}

type PushStreamingInfo struct {
	ID             string     `json:"id"`
	State          string     `json:"state"`
	Vhost          string     `json:"vhost"`
	App            string     `json:"app"`
	Stream         StreamInfo `json:"stream"`
	Protocol       string     `json:"protocol"`
	URL            string     `json:"url"`
	StreamKey      string     `json:"streamKey"`
	SentBytes      int        `json:"sentBytes"`
	SentTime       int        `json:"sentTime"`
	Sequence       int        `json:"sequence"`
	TotalSentBytes int        `json:"totalsentBytes"`
	TotalSentTime  int        `json:"totalsentTime"`
	CreatedTime    time.Time  `json:"createdTime"`
	StartTime      time.Time  `json:"startTime"`
	FinishTime     time.Time  `json:"finishTime"`
}

func (c *omePushApiClient) List(ctx context.Context) (*[]PushStreamingInfo, error) {
	p := fmt.Sprintf("%s:pushes", c.getBaseUrl())

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := c.handleError(ctx, &resp); err != nil {
		return nil, err
	}

	var parsedResp PushApiCommonResponse[[]PushStreamingInfo]
	if err := custhttp.JSONResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return parsedResp.Response, nil
}
