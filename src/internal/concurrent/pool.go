package custcon

import (
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"log"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

func New(size int) *ants.Pool {
	pool, err := ants.NewPool(
		size,
		ants.WithPreAlloc(true),
		ants.WithNonblocking(false),
		ants.WithLogger(logger.NewZapToAntsLogger(zap.L())),
	)
	if err != nil {
		log.Fatalf("pool.New: err = %s", err)
	}
	return pool
}
