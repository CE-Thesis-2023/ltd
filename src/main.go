package main

import (
	"context"
	"go.uber.org/zap"
	eventsapi "labs/local-transcoder/api/events"
	publicapi "labs/local-transcoder/api/public"
	"labs/local-transcoder/biz/service"
	"labs/local-transcoder/helper/factory"
	"labs/local-transcoder/internal/app"
	"labs/local-transcoder/internal/cache"
	"labs/local-transcoder/internal/configs"
	custdb "labs/local-transcoder/internal/db"
	custhttp "labs/local-transcoder/internal/http"
	"labs/local-transcoder/internal/logger"
	custmqtt "labs/local-transcoder/internal/mqtt"
	"labs/local-transcoder/models/db"
	"time"
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
				app.WithFactoryHook(func() error {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					custdb.Init(
						context.Background(),
						custdb.WithGlobalConfigs(&configs.Sqlite),
					)
					custdb.Migrate(&db.Camera{})

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
					return nil
				}),
				app.WithShutdownHook(func(ctx context.Context) {
					custdb.Stop(ctx)
					custmqtt.StopClient(ctx)
					service.Shutdown()
					logger.Close()
				}),
			}
		},
	)
}
