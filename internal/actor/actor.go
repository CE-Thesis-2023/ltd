package custactors

import (
	"labs/local-transcoder/internal/logger"
	"sync"

	"github.com/anthdm/hollywood/actor"
	"go.uber.org/zap"
)

var (
	pool *actor.Engine
	once sync.Once
)

func Init() {
	once.Do(func() {
		p, err := actor.NewEngine(nil)
		if err != nil {
			logger.SFatal("actor.Init error", zap.Error(err))
			return
		}

		pool = p
	})
}

func Pool() *actor.Engine {
	return pool
}
