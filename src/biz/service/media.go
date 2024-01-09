package service

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/helper"
	"github.com/CE-Thesis-2023/ltd/helper/factory"
	"github.com/CE-Thesis-2023/ltd/internal/cache"
	custcon "github.com/CE-Thesis-2023/ltd/internal/concurrent"
	custdb "github.com/CE-Thesis-2023/ltd/internal/db"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	"github.com/CE-Thesis-2023/ltd/internal/ome"
	"github.com/CE-Thesis-2023/ltd/models/db"
	"github.com/CE-Thesis-2023/ltd/models/events"
	"github.com/CE-Thesis-2023/ltd/models/ms"
	"github.com/CE-Thesis-2023/ltd/models/rest"
	"os/exec"

	"github.com/dgraph-io/ristretto"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type onGoingProcess struct {
	SourceUrl      string
	DestinationUrl string
	proc           *exec.Cmd
}

func (c *onGoingProcess) Cancel(ctx context.Context) error {
	if c.proc != nil {
		return c.proc.Cancel()
	}
	return nil
}

type mediaService struct {
	db               *custdb.LayeredDb
	omeClient        ome.OmeClientInterface
	streamingPool    *ants.Pool
	cache            *ristretto.Cache
	onGoingProcesses map[string]*onGoingProcess
}

func newMediaService() MediaServiceInterface {
	return &mediaService{
		db:               custdb.Layered(),
		omeClient:        factory.Ome(),
		streamingPool:    custcon.New(20),
		cache:            cache.Cache(),
		onGoingProcesses: map[string]*onGoingProcess{},
	}
}

func (s mediaService) Shutdown() {
	logger.SInfo("mediaService.Shutdown: shutdown received")
	for cameraId, p := range s.onGoingProcesses {
		if p.proc != nil {
			if err := p.proc.Cancel(); err != nil {
				logger.SDebug("mediaService.Shutdown: cancel process", zap.Error(err))
				continue
			}
		}
		delete(s.onGoingProcesses, cameraId)
		logger.SDebug("mediaService.Shutdown: canceled stream", zap.String("cameraId", cameraId))
	}
	s.streamingPool.Release()
	logger.SDebug("mediaService.Shutdown: released streaming pool")
}

type MediaServiceInterface interface {
	AdmissionWebhook(ctx context.Context, req *ms.AdmissionWebhookRequest) (*ms.AdmissionWebhookResponse, error)
	RequestPullRtsp(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error
	RequestPushSrt(ctx context.Context, req *ms.PushStreamingRequest) (*ome.StartPushStreamingResponse, error)
	RequestFFmpegRtspToSrt(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error
	CancelFFmpegRtspToSrt(ctx context.Context, camera *db.Camera) error
	ListOngoingStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error)
	Shutdown()
}

func (s *mediaService) AdmissionWebhook(ctx context.Context, req *ms.AdmissionWebhookRequest) (*ms.AdmissionWebhookResponse, error) {
	if s.isOutgoing(req) {
		logger.SInfo("s.AdmissionWebhook: allow all",
			zap.String("direction", string(req.Request.Direction)),
			zap.String("status", string(req.Request.Status)),
		)
		if s.isClosing(req) {
			return &ms.AdmissionWebhookResponse{
				Allowed:  helper.Bool(true),
				NewURL:   helper.String(req.Request.NewURL),
				Lifetime: helper.Int(0), // infinity
				Reason:   helper.String("allows all outgoing admission"),
			}, nil
		}
		return &ms.AdmissionWebhookResponse{}, nil
	}

	if !s.isAllowedToUseThisTranscoder(req) {
		logger.SInfo("s.AdmissionWebhook: not allowed", zap.String("ip", req.Client.Address))
		return &ms.AdmissionWebhookResponse{
			Allowed:  helper.Bool(false),
			NewURL:   helper.String(""),
			Lifetime: helper.Int(0),
			Reason:   helper.String("unauthorized"),
		}, nil
	}

	if err := s.startSrtPushStreaming(ctx, req); err != nil {
		logger.SInfo("s.AdmissionWebhook: not allowed", zap.Error(err))
		return &ms.AdmissionWebhookResponse{
			Allowed:  helper.Bool(false),
			NewURL:   helper.String(""),
			Lifetime: helper.Int(0),
			Reason:   helper.String("unable to start push streaming"),
		}, nil
	}

	logger.SInfo("s.AdmissionWebhook: allowed", zap.String("ip", req.Client.Address))
	return &ms.AdmissionWebhookResponse{
		Allowed:  helper.Bool(true),
		NewURL:   &req.Request.NewURL,
		Lifetime: helper.Int(0),
		Reason:   helper.String("authorized"),
	}, nil
}

func (s *mediaService) isOutgoing(req *ms.AdmissionWebhookRequest) bool {
	return req.Request.Direction == ms.DirectionOutgoing
}

func (s *mediaService) isClosing(req *ms.AdmissionWebhookRequest) bool {
	return req.Request.Status == ms.StatusClosing
}

func (s *mediaService) isAllowedToUseThisTranscoder(req *ms.AdmissionWebhookRequest) bool {
	return true
}

func (s *mediaService) startSrtPushStreaming(ctx context.Context, req *ms.AdmissionWebhookRequest) error {
	logger.SDebug("startSrtPushStreaming", zap.String("ip", req.Client.Address))
	return nil
}

func (s *mediaService) ListOngoingStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error) {
	logger.SDebug("ListOngoingStreams: request received")
	streams := s.onGoingProcesses
	resp := &rest.DebugListStreamsResponse{}
	for cameraId, s := range streams {
		resp.Streams = append(resp.Streams, rest.StreamInfo{
			CameraId:       cameraId,
			SourceUrl:      s.SourceUrl,
			DestinationUrl: s.DestinationUrl,
		})
	}
	logger.SDebug("ListOngoingStreams: streams", zap.Any("streams", resp))
	return resp, nil
}
