package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custcron "github.com/CE-Thesis-2023/ltd/src/internal/cron"
	"github.com/CE-Thesis-2023/ltd/src/internal/events"
	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/ltd/src/internal/mqtt"
	"github.com/CE-Thesis-2023/ltd/src/internal/ws"

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

	for _, s := range opts.httpServers {
		s := s
		go func() {
			logger.Infof("Run: start HTTP server name = %s", s.Name())
			if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Infof("Run: start HTTP server err = %s", err)
			}
		}()
	}

	for _, s := range opts.natsServers {
		s := s
		go func() {
			logger.Info("Run: start embedded NATS server", zap.String("name", s.Name()))
			if err := s.Start(); err != nil {
				logger.Infof("Run: start embedded NATS server error", zap.Error(err))
			}
		}()
	}

	for _, s := range opts.mqttServers {
		s := s
		go func() {
			logger.Infof("Run: start embedded MQTT server", zap.String("name", s.Name()))
			if err := s.Start(); err != nil {
				logger.Infof("Run: start embedded MQTT server error", zap.Error(err))
			}
		}()
	}

	for _, s := range opts.schedulers {
		s := s
		go func() {
			logger.Infof("Run: start scheduler", zap.String("name", s.Name()))
			if err := s.Start(); err != nil {
				logger.Info("Run: start scheduler error", zap.Error(err))
			}
		}()
	}

	for _, s := range opts.webSocketClients {
		s := s
		go func() {
			logger.Infof("Run: start WebSocket client")
			if err := s.Run(); err != nil {
				logger.Info("Run: start WebSocket client error", zap.Error(err))
			}
		}()
	}

	if opts.factoryHook != nil {
		if err := opts.factoryHook(); err != nil {
			logger.Fatalf("Run: factoryHook err = %s", err)
			return
		}
	}

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if opts.shutdownHook != nil {
		opts.shutdownHook(ctx)
	}

	var wg sync.WaitGroup

	wg.Add(5)

	go func() {
		defer wg.Done()
		for _, s := range opts.httpServers {
			s := s
			logger.Infof("Run: stop HTTP server name = %s", s.Name())
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, s := range opts.natsServers {
			s := s
			logger.Infof("Run: stop NATS embedded server name = %s", s.Name())
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, s := range opts.mqttServers {
			s := s
			logger.Infof("Run: stop MQTT embedded server name = %s", s.Name())
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, s := range opts.schedulers {
			s := s
			logger.Infof("Run: stop scheduler name = %s", s.Name())
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for _, s := range opts.webSocketClients {
			s := s
			logger.Infof("Run: stop WebSocket")
			if err := s.Stop(ctx); err != nil {
				log.Fatal(err)
			}
		}
	}()

	wg.Wait()

	zap.L().Sync()
	log.Print("Run: shutdown complete")
}

type RegistrationFunc func(configs *configs.Configs, logger *zap.Logger) []Optioner
type FactoryHook func() error
type ShutdownHook func(ctx context.Context)

type Options struct {
	httpServers      []*custhttp.HttpServer
	natsServers      []*events.EmbeddedNats
	mqttServers      []*custmqtt.EmbeddedMqtt
	schedulers       []*custcron.Scheduler
	webSocketClients []*ws.WebSocketClient

	factoryHook  FactoryHook
	shutdownHook ShutdownHook
}

type Optioner func(opts *Options)

func WithHttpServer(server *custhttp.HttpServer) Optioner {
	return func(opts *Options) {
		if server != nil {
			opts.httpServers = append(opts.httpServers, server)
		}
	}
}

func WithNatsServer(server *events.EmbeddedNats) Optioner {
	return func(opts *Options) {
		if server != nil {
			opts.natsServers = append(opts.natsServers, server)
		}
	}
}

func WithMqttServer(server *custmqtt.EmbeddedMqtt) Optioner {
	return func(opts *Options) {
		if server != nil {
			opts.mqttServers = append(opts.mqttServers, server)
		}
	}
}

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

func WithScheduling(scheduler *custcron.Scheduler) Optioner {
	return func(opts *Options) {
		opts.schedulers = append(opts.schedulers, scheduler)
	}
}

func WithWebSocketClient(client *ws.WebSocketClient) Optioner {
	return func(opts *Options) {
		opts.webSocketClients = append(opts.webSocketClients, client)
	}
}
