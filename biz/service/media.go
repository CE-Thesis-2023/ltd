package service

import (
	"context"
	"labs/local-transcoder/helper"
	"labs/local-transcoder/internal/configs"
	custdb "labs/local-transcoder/internal/db"
	"labs/local-transcoder/internal/logger"
	"labs/local-transcoder/internal/ome"
	"labs/local-transcoder/models/ms"

	"go.uber.org/zap"
)

type mediaService struct {
	db        *custdb.LayeredDb
	omeClient ome.OmeClientInterface
}

func newMediaService() MediaServiceInterface {
	localTranscoderOmeConfigs := configs.Get().LocalTranscoder

	omeClient := ome.NewOmeClient(&localTranscoderOmeConfigs)
	return &mediaService{
		db:        custdb.Layered(),
		omeClient: omeClient,
	}
}

type MediaServiceInterface interface {
	AdmissionWebhook(ctx context.Context, req *ms.AdmissionWebhookRequest) (*ms.AdmissionWebhookResponse, error)
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
