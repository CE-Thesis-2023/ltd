package main

import (
	"context"
	"go.uber.org/zap"
	eventsapi "github.com/CE-Thesis-2023/ltd/api/events"
	publicapi "github.com/CE-Thesis-2023/ltd/api/public"
	"github.com/CE-Thesis-2023/ltd/biz/service"
	"github.com/CE-Thesis-2023/ltd/helper/factory"
	"github.com/CE-Thesis-2023/ltd/internal/app"
	"github.com/CE-Thesis-2023/ltd/internal/cache"
	"github.com/CE-Thesis-2023/ltd/internal/configs"
	custdb "github.com/CE-Thesis-2023/ltd/internal/db"
	custhttp "github.com/CE-Thesis-2023/ltd/internal/http"
	"github.com/CE-Thesis-2023/ltd/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/ltd/internal/mqtt"
	"github.com/CE-Thesis-2023/ltd/models/db"
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
