package service

import (
	"context"
	"fmt"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	"github.com/CE-Thesis-2023/ltd/internal/ome"
	"github.com/CE-Thesis-2023/ltd/models/db"
	"github.com/CE-Thesis-2023/ltd/models/events"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

func (s *mediaService) RequestPullRtsp(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	omeListResp, err := s.omeClient.Streams().List(ctx)
	if err != nil {
		logger.SDebug("requestStream: list streams error", zap.Error(err))
		return err
	}

	for _, stream := range omeListResp.Names {
		if strings.EqualFold(stream, req.CameraId) {
			logger.SError("requestStream: stream already started", zap.String("name", req.CameraId))
			return custerror.ErrorAlreadyExists
		}
	}

	logger.SDebug("requestStream: stream not started, attempting to ask media server to enable stream")

	if err := s.startPullRtspStream(ctx, camera, req); err != nil {
		logger.SError("requestStream: startPullRtspStream error", zap.Error(err))
		return err
	}

	logger.SInfo("requestStream: success")
	return nil
}

func (s *mediaService) startPullRtspStream(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error {
	if err := s.omeClient.Streams().CreatePull(ctx, &ome.StreamCreationRequest{
		Name: req.CameraId,
		URLs: []string{
			s.buildRtspStreamUrl(camera, req),
		},
		Properties: ome.StreamProperties{
			Persistent:            true, // dont delete stream if no viewer or no input
			IgnoreRtcpSRTimestamp: false,
		},
	}); err != nil {
		logger.SDebug("startPullRtspStream: CreatePull error", zap.Error(err))
		return err
	}
	logger.SDebug("startPullRtspStream: success")
	return nil
}

func (s *mediaService) buildRtspStreamUrl(camera *db.Camera, req *events.CommandStartStreamInfo) string {
	u := &url.URL{}
	u.Scheme = "rtsp"
	u.Host = camera.Ip
	if camera.Port != 0 {
		u.Host = fmt.Sprintf("%s:%d", camera.Ip, camera.Port)
	}
	u = u.JoinPath("/ISAPI", "/Streaming", "channels", s.calculateChannelId(camera, req))
	u.User = url.UserPassword(camera.Username, camera.Password)
	url := u.String()
	logger.SDebug("buildRtspStreamUrl: stream url", zap.String("url", url))
	return url
}

func (s *mediaService) calculateChannelId(camera *db.Camera, req *events.CommandStartStreamInfo) string {
	channelId := req.ChannelId
	return fmt.Sprintf("%s01", channelId)
}
