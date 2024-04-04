package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"
)

func Run(shutdownTimeout time.Duration, registration RegistrationFunc) {
	ctx := context.Background()
	configs.Init(ctx)

	globalConfigs := configs.Get()

	loggerConfigs := globalConfigs.Logger
	logger.Init(ctx, logger.WithGlobalConfigs(&loggerConfigs))

	options := registration(globalConfigs, logger.Logger())

	opts := Options{}
	for _, optioner := range options {
		optioner(&opts)
	}

	logger := zap.L().Sugar()

	logger.Infof("Run: configs = %s", globalConfigs.String())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	reconcilerContext, cancel := context.WithCancel(context.Background())

	if opts.reconciler != nil {
		opts.reconciler(reconcilerContext)
	}

	if opts.factoryHook != nil {
		if err := opts.factoryHook(); err != nil {
			logger.Fatalf("Run: factoryHook err = %s", err)
			return
		}
	}

	<-quit
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if opts.shutdownHook != nil {
		opts.shutdownHook(ctx)
	}

	var wg sync.WaitGroup

	wg.Wait()

	zap.L().Sync()
	log.Print("Run: shutdown complete")
}

type RegistrationFunc func(configs *configs.Configs, logger *zap.Logger) []Optioner
type FactoryHook func() error
type ShutdownHook func(ctx context.Context)

type Options struct {
	factoryHook  FactoryHook
	shutdownHook ShutdownHook
	reconciler   func(ctx context.Context)
}

type Optioner func(opts *Options)

func WithFactoryHook(cb FactoryHook) Optioner {
	return func(opts *Options) {
		opts.factoryHook = cb
	}
}

func WithShutdownHook(cb ShutdownHook) Optioner {
	return func(opts *Options) {
		opts.shutdownHook = cb
	}
}

func WithReconciler(r func(ctx context.Context)) Optioner {
	return func(opts *Options) {
		opts.reconciler = r
	}
}
