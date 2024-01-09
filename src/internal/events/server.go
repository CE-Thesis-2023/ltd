package events

import (
	"context"
	"crypto/tls"
	"github.com/CE-Thesis-2023/ltd/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	"log"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"go.uber.org/zap"
)

type EmbeddedNats struct {
	configs *Options

	server *server.Server
	name   string
}

func New(options ...Optioner) *EmbeddedNats {
	opts := Options{}
	for _, opt := range options {
		opt(&opts)
	}
	serverConfigs := opts.configs

	server := buildServer(&opts)

	if server == nil {
		return nil
	}

	return &EmbeddedNats{
		configs: &opts,
		server:  server,
		name:    serverConfigs.Name,
	}
}

func buildServer(options *Options) *server.Server {
	if options == nil {
		return nil
	}

	configs := options.configs
	if !configs.Enabled {
		return nil
	}

	serverOptions := server.Options{
		Host:                   configs.Host,
		Port:                   configs.Port,
		ServerName:             configs.Name,
		JetStream:              true,
		DisableJetStreamBanner: true,
	}

	if configs.HasAuth() {
		serverOptions.Username = configs.Username
		serverOptions.Password = configs.Password
	}

	if configs.Tls.Enabled() {
		serverTls, err := buildTlsConfigs(&configs.Tls)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		serverOptions.TLSConfig = serverTls
		serverOptions.TLS = true
	}

	server, err := server.NewServer(&serverOptions)
	if err != nil {
		log.Fatalf("buildServer: build server err = %s", err)
		return nil
	}

	if options.logger != nil {
		server.SetLoggerV2(logger.NewZapToNatsLogger(options.logger),
			true,
			false,
			false)
	} else {
		server.ConfigureLogger()
	}
	return server
}

func buildTlsConfigs(tlsConfigs *configs.TlsConfig) (*tls.Config, error) {
	configs, err := server.GenTLSConfig(&server.TLSConfigOpts{
		CertFile: tlsConfigs.Cert,
		KeyFile:  tlsConfigs.Key,
		CaFile:   tlsConfigs.Authority,
		Verify:   false,
		Insecure: true,
	})
	if err != nil {
		return nil, custerror.FormatInternalError("buildTlsConfigs: err = %s", err)
	}
	return configs, nil

}

type Options struct {
	configs *configs.EventStoreConfigs
	logger  *zap.SugaredLogger
}

type Optioner func(opts *Options)

func WithGlobalConfigs(configs *configs.EventStoreConfigs) Optioner {
	return func(opts *Options) {
		opts.configs = configs
	}
}

func WithZapLogger(logger *zap.SugaredLogger) Optioner {
	return func(opts *Options) {
		opts.logger = logger
	}
}

func (n *EmbeddedNats) Start() error {
	n.server.Start()
	if !n.server.ReadyForConnections(time.Second * 3) {
		return custerror.FormatInternalError("EmbeddedNats.Start: connection not ready")
	}
	return nil
}

func (n *EmbeddedNats) Stop(ctx context.Context) error {
	var dur time.Duration
	dl, ok := ctx.Deadline()
	if !ok {
		dur = 2 * time.Second
	} else {
		dur = time.Until(dl)
	}
	if n.server.ReadyForConnections(dur) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("EmbeddedNats.recover: panic caught err = %s", err)
			}
		}()
		n.server.Shutdown()
	}
	return nil
}

func (n *EmbeddedNats) Name() string {
	return n.name
}
