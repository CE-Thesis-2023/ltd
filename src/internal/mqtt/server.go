package custmqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	"log"
	"log/slog"
	"os"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"go.uber.org/zap"
)

type EmbeddedMqtt struct {
	options *Options

	server *mqtt.Server
}

func (s *EmbeddedMqtt) Start() error {
	if err := s.server.Serve(); err != nil {
		return custerror.FormatInternalError("EmbeddedMqtt.Start: err = %s", err)
	}
	return nil
}

func (s *EmbeddedMqtt) Stop(ctx context.Context) error {
	if err := s.server.Close(); err != nil {
		return custerror.FormatInternalError("EmbeddedMqtt.Stop: err = %s", err)
	}
	return nil
}

func (s *EmbeddedMqtt) Name() string {
	return s.options.globalConfigs.Name
}

func New(options ...Optioner) *EmbeddedMqtt {
	opts := Options{}
	for _, o := range options {
		o(&opts)
	}

	tcpListener, err := buildTcpListener(&opts)
	if err != nil {
		log.Fatalf("mqtt.buildTcpListener: err = %s", err)
		return nil
	}

	mqttOptions := buildMqttOptions(&opts)

	server := mqtt.New(mqttOptions)
	server.AddListener(tcpListener)

	globalConfigs := opts.globalConfigs
	if globalConfigs.HasAuth() {
		server.AddHook(&auth.Hook{}, auth.Options{
			Ledger: &auth.Ledger{
				Auth: auth.AuthRules{
					auth.AuthRule{
						Username: auth.RString(globalConfigs.Username),
						Password: auth.RString(globalConfigs.Password),
						Allow:    true,
					},
				},
			},
		})
	} else {
		server.AddHook(&auth.AllowHook{}, nil)
	}

	return &EmbeddedMqtt{
		server:  server,
		options: &opts,
	}
}

func buildTcpListener(options *Options) (*listeners.TCP, error) {
	globalConfigs := options.globalConfigs

	addr := fmt.Sprintf("%s:%d", globalConfigs.Host, globalConfigs.Port)

	listenerConfigs := &listeners.Config{}

	tlsConfigs, err := makeTlsConfigs(globalConfigs)
	if err != nil {
		return nil, custerror.FormatInternalError("mqtt.buildTcpListener: err = %s", err)
	}

	if tlsConfigs != nil {
		listenerConfigs.TLSConfig = tlsConfigs
	}

	tcp := listeners.NewTCP(globalConfigs.Name, addr, listenerConfigs)

	return tcp, nil
}

func buildMqttOptions(options *Options) *mqtt.Options {
	globalConfigs := options.globalConfigs
	logLevel := globalConfigs.Level

	mqttOptions := &mqtt.Options{}

	if options.logger != nil {
		slogger := slog.New(logger.NewZapToSlogHandler(options.logger))
		mqttOptions.Logger = slogger
	} else {
		if len(logLevel) > 0 {
			slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: parseLogLevel(logLevel),
			}))
			mqttOptions.Logger = slogger
		}
	}

	return mqttOptions
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "error":
		return slog.LevelError
	case "info":
		return slog.LevelInfo
	default:
		log.Printf("mqtt.parseLogLevel: log level unknown, fallback to 'info' level = %s", level)
		return slog.LevelInfo
	}
}

func makeTlsConfigs(globalConfigs *configs.EventStoreConfigs) (*tls.Config, error) {
	tlsConfigs := globalConfigs.Tls
	if !tlsConfigs.Enabled() {
		return nil, nil
	}

	credentials, err := tls.LoadX509KeyPair(tlsConfigs.Cert, tlsConfigs.Key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		InsecureSkipVerify: true,
		Certificates: []tls.Certificate{
			credentials,
		},
	}, nil
}

type Options struct {
	globalConfigs *configs.EventStoreConfigs
	logger        *zap.Logger
}

type Optioner func(options *Options)

func WithGlobalConfigs(configs *configs.EventStoreConfigs) Optioner {
	return func(options *Options) {
		options.globalConfigs = configs
	}
}

func WithZapLogger(logger *zap.Logger) Optioner {
	return func(options *Options) {
		options.logger = logger
	}
}
