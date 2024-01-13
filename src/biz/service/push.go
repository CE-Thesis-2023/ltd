package service

import (
	"context"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/models/ms"
	"net/url"
)

func (s *mediaService) buildPushSrtUrl(ctx context.Context, req *ms.PushStreamingRequest) string {
	configs := configs.Get().CloudMediaServer

	streamUrl := &url.URL{}
	streamUrl.Scheme = "srt"
	streamUrl.Host = configs.Host
	if configs.Port != 0 {
		streamUrl.Host = fmt.Sprintf("%s:%d", configs.Host, configs.Port)
	}
	streamUrl = streamUrl.JoinPath(req.StreamName)
	encodedStreamId := streamUrl.String()

	queries := streamUrl.Query()
	queries.Add("streamid", encodedStreamId)
	streamUrl.RawQuery = queries.Encode()

	url := streamUrl.String()
	return url
}
