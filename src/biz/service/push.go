package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/src/helper"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/internal/ome"
	"github.com/CE-Thesis-2023/ltd/src/models/ms"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *mediaService) RequestPushSrt(ctx context.Context, req *ms.PushStreamingRequest) (*ome.StartPushStreamingResponse, error) {
	logger.SDebug("ms.RequestPushSrt", zap.Any("request", req))

	existing, err := s.getPushStreamByName(ctx, req)
	if err != nil {
		if !errors.Is(err, custerror.ErrorNotFound) {
			logger.SError("RequestPushSrt: getPushStreamByName",
				zap.Error(err))
			return nil, err
		}
	}
	if existing != nil {
		logger.SDebug("RequestPushSrt: push stream already exists",
			zap.Any("pushStream", existing))
		return nil, custerror.ErrorAlreadyExists
	}

	url := s.buildPushSrtUrl(ctx, req)
	logger.SDebug("ms.RequestPushSrt", zap.String("url", url))

	resp, err := s.omeClient.Push().Start(ctx, &ome.StartPushStreamingRequest{
		ID: uuid.NewString(),
		Stream: ome.OriginStreamInfo{
			Name:         req.StreamName,
			VariantNames: []string{},
		},
		Protocol:  ome.ProtocolSRT,
		URL:       url,
		StreamKey: helper.String(""),
	})

	if err != nil {
		logger.SError("RequestPushSrt: Push.Start", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

func (s *mediaService) buildPushSrtUrl(ctx context.Context, req *ms.PushStreamingRequest) string {
	configs := configs.Get().CloudMediaServer

	streamUrl := &url.URL{}
	streamUrl.Scheme = "srt"
	streamUrl.Host = configs.Host
	if configs.Port != 0 {
		streamUrl.Host = fmt.Sprintf("%s:%d", configs.Host, configs.Port)
	}
	streamUrl = streamUrl.JoinPath(configs.ApplicationName, req.StreamName)
	encodedStreamId := streamUrl.String()

	queries := streamUrl.Query()
	queries.Add("mode", "caller")
	queries.Add("streamid", encodedStreamId)
	streamUrl.RawQuery = queries.Encode()

	url := streamUrl.String()
	return url
}

func (s *mediaService) getPushStreamByName(ctx context.Context, req *ms.PushStreamingRequest) (*ome.PushStreamingInfo, error) {
	pushes, err := s.omeClient.Push().List(ctx)
	if err != nil {
		logger.SDebug("getPushStreamByName: Push.List error", zap.Error(err))
		return nil, err
	}
	for _, pushStream := range *pushes {
		if strings.EqualFold(pushStream.Stream.Name, req.StreamName) {
			return &pushStream, nil
		}
	}
	logger.SDebug("getPushStreamByName: push stream not found", zap.String("name", req.StreamName))
	return nil, custerror.ErrorNotFound
}
