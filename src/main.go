package main

import (
	"context"
	"time"

	eventsapi "github.com/CE-Thesis-2023/ltd/src/api/events"
	"github.com/CE-Thesis-2023/ltd/src/biz/reconciler"
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/helper/factory"
	"github.com/CE-Thesis-2023/ltd/src/internal/app"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/ltd/src/internal/mqtt"
	"go.uber.org/zap"
)

func main() {
	app.Run(
		time.Second*10,
		func(configs *configs.Configs, zl *zap.Logger) []app.Optioner {
			return []app.Optioner{
				app.WithReconciler(reconciler.
					GetReconciler().
					Run),
				app.WithFactoryHook(func() error {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					factory.Init(ctx, configs)
					service.Init()
					reconciler.Init()
					eventsapi.Init(ctx)

					custmqtt.InitClient(
						context.Background(),
						custmqtt.WithClientGlobalConfigs(&configs.MqttStore),
						custmqtt.WithOnReconnection(eventsapi.Register),
						custmqtt.WithOnConnectError(func(err error) {
							logger.Error("MQTT connection failed",
								zap.Error(err))
						}),
						custmqtt.WithClientError(eventsapi.ClientErrorHandler),
						custmqtt.WithOnServerDisconnect(eventsapi.DisconnectHandler),
						custmqtt.WithHandlerRegister(eventsapi.RouterHandler()),
					)
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
