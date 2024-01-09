package factory

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/internal/ome"
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
