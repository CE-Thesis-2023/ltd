package main

import (
	"context"
	eventsapi "labs/local-transcoder/api/events"
	"labs/local-transcoder/helper/factory"
	"labs/local-transcoder/models/db"

	// privateapi "labs/local-transcoder/api/private"
	publicapi "labs/local-transcoder/api/public"
	"labs/local-transcoder/biz/service"
	custactors "labs/local-transcoder/internal/actor"
	"labs/local-transcoder/internal/app"
	"labs/local-transcoder/internal/cache"
	"labs/local-transcoder/internal/configs"
	custmqtt "labs/local-transcoder/internal/mqtt"

	// custcron "labs/local-transcoder/internal/cron"
	custdb "labs/local-transcoder/internal/db"
	custhttp "labs/local-transcoder/internal/http"
	"labs/local-transcoder/internal/logger"

	// custmqtt "labs/local-transcoder/internal/mqtt"
	"time"

	// "github.com/go-co-op/gocron"
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
				app.WithFactoryHook(func() error {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					custdb.Init(
						context.Background(),
						custdb.WithGlobalConfigs(&configs.Sqlite),
					)
					custdb.Migrate(&db.Camera{})

					cache.Init()
					custactors.Init()
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
					logger.Close()
				}),
			}
		},
	)
}
