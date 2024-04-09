package app

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/reconciler"
	"github.com/CE-Thesis-2023/ltd/src/service"
	"go.uber.org/zap"
)

func Run() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second)
	defer cancel()
	configs.Init(ctx)

	globalConfigs := configs.Get()
	logger.Init(
		ctx,
		logger.WithGlobalConfigs(&globalConfigs.Logger))

	logger.SInfo("application starting",
		zap.Reflect("configs", globalConfigs))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	hikvisionClient, err := hikvision.NewClient(
		hikvision.WithPoolSize(10),
	)
	if err != nil {
		logger.SFatal("failed to create hikvision client", zap.Error(err))
	}
	controlPlaneService := service.NewControlPlaneService(&globalConfigs.DeviceInfo)
	commandService := service.NewCommandService(hikvisionClient)
	mediaService := service.NewMediaService()

	mediaController := service.NewMediaController(mediaService)

	reconciler := reconciler.NewReconciler(
		controlPlaneService,
		&globalConfigs.DeviceInfo,
		commandService,
		mediaController,
	)

	reconcilerContext, reconcilerCancel := context.WithCancel(context.Background())
	if reconciler != nil {
		go func() {
			reconciler.Run(reconcilerContext)
			reconcilerCancel()
		}()
	}

	defer func() {
		ctx, cancel = context.WithTimeout(
			context.Background(),
			5*time.Second)
		defer cancel()
		logger.SInfo("application shutdown complete")
	}()

	<-quit
	logger.SInfo("application shutdown requested")
	reconcilerCancel()
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
