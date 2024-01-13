package factory

import (
	"context"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"sync"

	"go.uber.org/zap"
)

var once sync.Once

var (
	hikvisionClient hikvision.Client
)

func Init(ctx context.Context, configs *configs.Configs) {
	once.Do(func() {
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

func Hikvision() hikvision.Client {
	return hikvisionClient
}
