package ome

import (
	"context"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"
	custhttp "github.com/CE-Thesis-2023/ltd/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
)

type OmeStreamsApiInterface interface {
	Get(ctx context.Context, streamName string) (*GetStreamInfoResponse, error)
	CreatePull(ctx context.Context, req *StreamCreationRequest) error
	List(ctx context.Context) (*ListStreamsResponse, error)
	Delete(ctx context.Context, streamName string) error
}

type omeStreamsApiClient struct {
	localConfigs *configs.OvenMediaEngineConfigs
	restClient   fastshot.ClientHttpMethods
}

type StreamProperties struct {
	Persistent                    bool `json:"persistent"`
	NoInputFailoverTimeoutMs      int  `json:"noInputFailoverTimeoutMs"`
	UnusedStreamDeletionTimeoutMs int  `json:"unusedStreamDeletionTimeoutMs"`
	IgnoreRtcpSRTimestamp         bool `json:"ignoreRtcpSRTimestamp"`
}

type StreamCreationRequest struct {
	Name       string           `json:"name"`
	URLs       []string         `json:"urls"`
	Properties StreamProperties `json:"properties"`
}

func (c *omeStreamsApiClient) getBaseUrl() string {
	return fmt.Sprintf("/v1/vhosts/%s/apps/%s/streams",
		c.localConfigs.VirtualHostName,
		c.localConfigs.ApplicationName)
}

type StreamsApiCommonResponse[T interface{}] struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Response   *T     `json:"response,omitempty"`
}

func (c *omeStreamsApiClient) handleError(ctx context.Context, resp *fastshot.Response) error {
	switch resp.StatusCode() {
	case 400:
		var parsedResp StreamsApiCommonResponse[interface{}]
		if err := custhttp.JSONResponse(resp, &parsedResp); err != nil {
			return err
		}
		return custerror.FormatInvalidArgument(parsedResp.Message)
	case 401:
		return custerror.ErrorPermissionDenied
	case 404:
		return custerror.ErrorNotFound
	case 409:
		var parsedResp StreamsApiCommonResponse[interface{}]
		if err := custhttp.JSONResponse(resp, &parsedResp); err != nil {
			return err
		}
		return custerror.FormatAlreadyExists(parsedResp.Message)
	case 500, 502:
		return custerror.ErrorInternal
	}

	return nil
}

func (c *omeStreamsApiClient) CreatePull(ctx context.Context, req *StreamCreationRequest) error {
	p := c.getBaseUrl()
	resp, err := c.restClient.POST(p).
		Context().Set(ctx).
		Body().AsJSON(req).Send()
	if err != nil {
		return err
	}

	if err := c.handleError(ctx, &resp); err != nil {
		return err
	}

	return nil
}

type Video struct {
	Bitrate          string  `json:"bitrate"`
	Bypass           bool    `json:"bypass"`
	Codec            string  `json:"codec"`
	Framerate        float64 `json:"framerate"`
	HasBframes       bool    `json:"hasBframes"`
	KeyFrameInterval int     `json:"keyFrameInterval"`
	Height           int     `json:"height"`
	Width            int     `json:"width"`
}

type Audio struct {
	Bitrate    string `json:"bitrate"`
	Bypass     bool   `json:"bypass"`
	Channel    int    `json:"channel"`
	Codec      string `json:"codec"`
	Samplerate int    `json:"samplerate"`
}

type Track struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Video *Video `json:"video,omitempty"`
	Audio *Audio `json:"audio,omitempty"`
}

type Input struct {
	CreatedTime string  `json:"createdTime"`
	SourceType  string  `json:"sourceType"`
	Tracks      []Track `json:"tracks"`
}

type Output struct {
	Name   string  `json:"name"`
	Tracks []Track `json:"tracks"`
}

type GetStreamInfoResponse struct {
	Input   Input    `json:"input"`
	Name    string   `json:"name"`
	Outputs []Output `json:"outputs"`
}

func (c *omeStreamsApiClient) Get(ctx context.Context, streamName string) (*GetStreamInfoResponse, error) {
	p := fmt.Sprintf("%s/%s", c.getBaseUrl(), streamName)

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := c.handleError(ctx, &resp); err != nil {
		return nil, err
	}

	var parsedResp StreamsApiCommonResponse[GetStreamInfoResponse]
	if err := custhttp.JSONResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}
	return parsedResp.Response, nil
}

type ListStreamsResponse struct {
	Names []string
}

func (c *omeStreamsApiClient) List(ctx context.Context) (*ListStreamsResponse, error) {
	p := c.getBaseUrl()

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := c.handleError(ctx, &resp); err != nil {
		return nil, err
	}

	var parsedResp StreamsApiCommonResponse[[]string]
	if err := custhttp.JSONResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &ListStreamsResponse{
		Names: *parsedResp.Response,
	}, nil
}

func (c *omeStreamsApiClient) Delete(ctx context.Context, streamName string) error {
	p := fmt.Sprintf("%s/%s", c.getBaseUrl(), streamName)
	resp, err := c.restClient.DELETE(p).
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
