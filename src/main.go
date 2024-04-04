package main

import (
	"context"
	"errors"
	"time"

	eventsapi "github.com/CE-Thesis-2023/ltd/src/api/events"
	publicapi "github.com/CE-Thesis-2023/ltd/src/api/public"
	wsapi "github.com/CE-Thesis-2023/ltd/src/api/websocket"
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/helper/factory"
	"github.com/CE-Thesis-2023/ltd/src/internal/app"
	"github.com/CE-Thesis-2023/ltd/src/internal/cache"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/ltd/src/internal/mqtt"
	"github.com/CE-Thesis-2023/ltd/src/internal/ws"
	"go.uber.org/zap"
)

func main() {
	app.Run(
		time.Second*10,
		func(configs *configs.Configs, zl *zap.Logger) []app.Optioner {
			return []app.Optioner{
				app.WithHttpServer(custhttp.New(
					custhttp.WithGlobalConfigs(&configs.Public),
					custhttp.WithErrorHandler(custhttp.GlobalErrorHandler()),
					custhttp.WithRegistration(publicapi.ServiceRegistration()),
					custhttp.WithMiddleware(custhttp.CommonPublicMiddlewares(&configs.Public)...),
				)),
				app.WithWebSocketClient(ws.NewWebSocketClient(
					ws.WithDeviceId(configs.DeviceInfo.DeviceId),
					ws.WithGlobalConfigs(&configs.WebSocket),
					ws.WithMessageHandler(wsapi.GetStandardHandler()),
				)),
				app.WithFactoryHook(func() error {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					cache.Init()
					factory.Init(ctx, configs)

					service.Init()
					eventsapi.Init(ctx)

					custmqtt.InitClient(
						context.Background(),
						custmqtt.WithClientGlobalConfigs(&configs.MqttStore),
						custmqtt.WithOnReconnection(eventsapi.Register),
						custmqtt.WithOnConnectError(func(err error) {
							logger.Error("MQTT Connection failed", zap.Error(err))
						}),
						custmqtt.WithClientError(eventsapi.ClientErrorHandler),
						custmqtt.WithOnServerDisconnect(eventsapi.DisconnectHandler),
						custmqtt.WithHandlerRegister(eventsapi.RouterHandler()),
					)

					if err := service.GetCommandService().RegisterDevice(ctx); err != nil {
						if !errors.Is(err, custerror.ErrorAlreadyExists) {
							logger.SDebug("RegisterDevice: error", zap.Error(err))
							return nil
						}
						logger.SDebug("RegisterDevice: device already registered")
					}

					if err := service.GetCommandService().UpdateCameraList(ctx); err != nil {
						logger.SError("UpdateCameraList: error", zap.Error(err))
						return err
					}

					if err := service.GetCommandService().StartAllEnabledStreams(ctx); err != nil {
						logger.SError("StartAllEnabledStreams: error", zap.Error(err))
					}
					return nil
				}),
				app.WithShutdownHook(func(ctx context.Context) {
					custmqtt.StopClient(ctx)
					service.Shutdown()
					logger.Close()
				}),
			}
		},
	)
}
