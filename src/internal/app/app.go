package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/hikvision"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/reconciler"
	"github.com/CE-Thesis-2023/ltd/src/service"
	"github.com/CE-Thesis-2023/ltd/src/sidecar"
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
	processorController := service.NewProcessorController(&configs.Get().OpenGate, mediaService)

	reconciler := reconciler.NewReconciler(
		controlPlaneService,
		&globalConfigs.DeviceInfo,
		commandService,
		mediaController,
		processorController,
	)
	sidecar := sidecar.NewHttpSidecar(commandService, reconciler)

	var wg sync.WaitGroup

	reconcilerContext, reconcilerCancel := context.WithCancel(context.Background())
	if reconciler != nil {
		wg.Add(1)
		go func() {
			reconciler.Run(reconcilerContext)
			reconcilerCancel()
			quit <- os.Interrupt
			defer wg.Done()
		}()
	}

	if sidecar != nil {
		wg.Add(1)
		go func() {
			if err := sidecar.Start(); err != nil {
				logger.SError("failed to start sidecar", zap.Error(err))
			}
			quit <- os.Interrupt
			defer wg.Done()
		}()
	}

	<-quit
	logger.SInfo("application shutdown requested")
	if sidecar != nil {
		if err := sidecar.Stop(ctx); err != nil {
			logger.SError("failed to stop sidecar", zap.Error(err))
		}
	}
	reconcilerCancel()
	wg.Wait()
	logger.SInfo("application shutdown complete")
}

type RegistrationFunc func(configs *configs.Configs, logger *zap.Logger) []Optioner
type FactoryHook func(ctx context.Context) error
type ShutdownHook func(ctx context.Context)

type Options struct {
	factoryHook       FactoryHook
	shutdownHook      ShutdownHook
	reconcilerFactory func() reconciler.Reconciler
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

func WithReconcilerFactory(r func() reconciler.Reconciler) Optioner {
	return func(opts *Options) {
		opts.reconcilerFactory = r
	}
}
