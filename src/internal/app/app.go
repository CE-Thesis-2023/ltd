package app

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/biz/reconciler"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"
)

func Run(shutdownTimeout time.Duration, registration RegistrationFunc) {
	ctx := context.Background()
	configs.Init(ctx)

	globalConfigs := configs.Get()
	configsBytes, _ := json.Marshal(globalConfigs)

	loggerConfigs := globalConfigs.Logger
	logger.Init(ctx, logger.WithGlobalConfigs(&loggerConfigs))

	options := registration(
		globalConfigs,
		logger.Logger())

	opts := Options{}
	for _, optioner := range options {
		optioner(&opts)
	}

	logger := zap.L()
	logger.Info("application starting",
		zap.Any("configs", json.RawMessage(configsBytes)))

	if opts.factoryHook != nil {
		if err := opts.factoryHook(ctx); err != nil {
			logger.Fatal("failed to run factory hook",
				zap.Error(err))
			return
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	reconcilerContext, cancel := context.WithCancel(context.Background())

	var reconciler reconciler.BaseReconciler
	if opts.reconcilerFactory != nil {
		reconciler = opts.reconcilerFactory()
	}

	if reconciler != nil {
		go reconciler.Run(reconcilerContext)
	}

	<-quit
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if opts.shutdownHook != nil {
		opts.shutdownHook(ctx)
	}

	zap.L().
		Sync()
	logger.Info("application shutdown complete")
}

type RegistrationFunc func(configs *configs.Configs, logger *zap.Logger) []Optioner
type FactoryHook func(ctx context.Context) error
type ShutdownHook func(ctx context.Context)

type Options struct {
	factoryHook       FactoryHook
	shutdownHook      ShutdownHook
	reconcilerFactory func() reconciler.BaseReconciler
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

func WithReconcilerFactory(r func() reconciler.BaseReconciler) Optioner {
	return func(opts *Options) {
		opts.reconcilerFactory = r
	}
}
