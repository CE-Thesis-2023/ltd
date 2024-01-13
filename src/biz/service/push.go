package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/ms"
	"go.uber.org/zap"
)

func (s *mediaService) buildPushSrtUrl(ctx context.Context, req *ms.PushStreamingRequest) string {
	configs := configs.Get().CloudMediaServer

	streamUrl := &url.URL{}
	streamUrl.Scheme = "srt"
	streamUrl.Host = configs.Host
	if configs.PublishPorts.Srt != 0 {
		streamUrl.Host = fmt.Sprintf("%s:%d", configs.Host, configs.PublishPorts.Srt)
	}

	queries := streamUrl.Query()
	queries.Add("streamid", fmt.Sprintf("publish:%s", req.StreamName))
	rawQuery, err := url.QueryUnescape(queries.Encode())
	logger.SError("buildPushSrtUrl: err = %s", zap.Error(err))
	streamUrl.RawQuery = rawQuery

	url := streamUrl.String()
	return url
}
