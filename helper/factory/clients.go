package factory

import (
	"context"
	"labs/local-transcoder/internal/configs"
	"labs/local-transcoder/internal/hikvision"
	"labs/local-transcoder/internal/logger"
	"labs/local-transcoder/internal/ome"
	"sync"

	"go.uber.org/zap"
)

var once sync.Once

var (
	omeClient       ome.OmeClientInterface
	hikvisionClient hikvision.Client
)

func Init(ctx context.Context, configs *configs.Configs) {
	once.Do(func() {
		omeClient = ome.NewOmeClient(&configs.LocalTranscoder)
		hikvi, err := hikvision.NewClient(
			hikvision.WithPoolSize(20),
		)
		if err != nil {
			logger.SFatal("factory.Init: hikvision.Client", zap.Error(err))
			return
		}
		hikvisionClient = hikvi
	})
}

func Ome() ome.OmeClientInterface {
	return omeClient
}

func Hikvision() hikvision.Client {
	return hikvisionClient
}
