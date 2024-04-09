package custmqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"go.uber.org/zap"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func NewClient(ctx context.Context, options ...ClientOptioner) *autopaho.ConnectionManager {
	opts := &ClientOptions{}
	for _, opt := range options {
		opt(opts)
	}

	globalConfigs := opts.globalConfigs
	connUrl := url.URL{}
	if globalConfigs.TlsEnabled {
		connUrl.Scheme = "tls"
	} else {
		connUrl.Scheme = "mqtt"
	}
	hostname := globalConfigs.Host

	if globalConfigs.Port > 0 {
		hostname = fmt.Sprintf("%s:%d", globalConfigs.Host, globalConfigs.Port)
	}
	connUrl.Host = hostname

	router := paho.NewStandardRouter()

	if opts.register != nil {
		opts.register(router)
	}

	clientConfigs := autopaho.ClientConfig{
		KeepAlive:         20,
		ConnectRetryDelay: time.Second * 5,
		ConnectTimeout:    time.Second * 2,
		BrokerUrls: []*url.URL{
			&connUrl,
		},
		ClientConfig: paho.ClientConfig{
			Router: router,
		},
	}

	if globalConfigs.TlsEnabled {
		tlsConfigs, err := makeTlsConfigs(globalConfigs)
		if err != nil {
			logger.SFatal("create TLS configuration failed", zap.Error(err))
			return nil
		}
		clientConfigs.TlsCfg = tlsConfigs
	}

	if globalConfigs.HasAuth() {
		clientConfigs.SetUsernamePassword(globalConfigs.Username, []byte(globalConfigs.Password))
	}

	if opts.reconCallback != nil {
		clientConfigs.OnConnectionUp = opts.reconCallback
	}

	if opts.connErrCallback != nil {
		clientConfigs.OnConnectError = opts.connErrCallback
	}

	if opts.clientErr != nil {
		clientConfigs.ClientConfig.OnClientError = opts.clientErr
	}

	if opts.serverDisconnect != nil {
		clientConfigs.ClientConfig.OnServerDisconnect = opts.serverDisconnect
	}

	connManager, err := autopaho.NewConnection(ctx, clientConfigs)
	if err != nil {
		logger.SFatal("MQTT connection failed",
			zap.Error(err))
		return nil
	}

	if err := connManager.AwaitConnection(ctx); err != nil {
		logger.SFatal("MQTT waiting for connection failed",
			zap.Error(err))
		return nil
	}

	return connManager
}

func makeTlsConfigs(globalConfigs *configs.EventStoreConfigs) (*tls.Config, error) {
	t := &tls.Config{
		InsecureSkipVerify: true,
	}
	return t, nil
}

type ClientOptions struct {
	globalConfigs    *configs.EventStoreConfigs
	reconCallback    func(cm *autopaho.ConnectionManager, connack *paho.Connack)
	connErrCallback  func(err error)
	serverDisconnect func(d *paho.Disconnect)
	clientErr        func(err error)
	register         RouterRegister
}

type ClientOptioner func(options *ClientOptions)

type RouterRegister func(router *paho.StandardRouter)

func WithClientGlobalConfigs(configs *configs.EventStoreConfigs) ClientOptioner {
	return func(options *ClientOptions) {
		options.globalConfigs = configs
	}
}

func WithOnReconnection(cb func(cm *autopaho.ConnectionManager, connack *paho.Connack)) ClientOptioner {
	return func(options *ClientOptions) {
		options.reconCallback = cb
	}
}

func WithOnConnectError(cb func(err error)) ClientOptioner {
	return func(options *ClientOptions) {
		options.connErrCallback = cb
	}
}

func WithOnServerDisconnect(cb func(d *paho.Disconnect)) ClientOptioner {
	return func(options *ClientOptions) {
		options.serverDisconnect = cb
	}
}

func WithClientError(cb func(err error)) ClientOptioner {
	return func(options *ClientOptions) {
		options.clientErr = cb
	}
}

func WithHandlerRegister(cb RouterRegister) ClientOptioner {
	return func(options *ClientOptions) {
		options.register = cb
	}
}
