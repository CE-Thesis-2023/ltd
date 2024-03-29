package service

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/src/internal/cache"
	custcon "github.com/CE-Thesis-2023/ltd/src/internal/concurrent"
	custdb "github.com/CE-Thesis-2023/ltd/src/internal/db"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/db"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"github.com/CE-Thesis-2023/ltd/src/models/rest"
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
	streamingPool    *ants.Pool
	cache            *ristretto.Cache
	onGoingProcesses map[string]*onGoingProcess
}

func newMediaService() MediaServiceInterface {
	return &mediaService{
		db:               custdb.Layered(),
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
	RequestFFmpegRtspToSrt(ctx context.Context, camera *db.Camera, req *events.CommandStartStreamInfo) error
	CancelFFmpegRtspToSrt(ctx context.Context, camera *db.Camera) error
	ListOngoingStreams(ctx context.Context) (*rest.DebugListStreamsResponse, error)
	Shutdown()
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
