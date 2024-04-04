package eventsapi

import (
	"context"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/src/helper"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	custmqtt "github.com/CE-Thesis-2023/ltd/src/internal/mqtt"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

func Register(cm *autopaho.ConnectionManager, connack *paho.Connack) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	subs := makeSubsciptions()
	if _, err := cm.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: subs,
	}); err != nil {
		logger.SError("unable to make MQTT subscriptions",
			zap.String("where", "api.events.Register"),
			zap.Reflect("subs", subs),
		)
		return
	}

	logger.SInfo("MQTT subscriptions made success", zap.Reflect("subs", subs))
}

func makeSubsciptions() []paho.SubscribeOptions {
	return []paho.SubscribeOptions{
		{Topic: fmt.Sprintf("commands/%s", configs.Get().DeviceInfo.DeviceId), QoS: 1},
		{Topic: fmt.Sprintf("ptzctrl/%s", configs.Get().DeviceInfo.DeviceId), QoS: 1},
	}
}

func ClientErrorHandler(err error) {
	logger := logger.Logger()

	logger.Error("MQTT Client", zap.Error(err))
}

func DisconnectHandler(d *paho.Disconnect) {
	logger := logger.Logger()

	logger.Error("MQTT Server Disconnect", zap.String("reason", d.Properties.ReasonString))
}

func RouterHandler() custmqtt.RouterRegister {
	return func(router *paho.StandardRouter) {
		handlers := GetStandardEventsHandler()
		router.RegisterHandler(
			fmt.Sprintf("commands/%s", configs.Get().DeviceInfo.DeviceId),
			WrapForHandlers(handlers.ReceiveRemoteCommands),
		)
		router.RegisterHandler(
			fmt.Sprintf("ptzctrl/%s", configs.Get().DeviceInfo.DeviceId),
			WrapForHandlers(handlers.ReceiveRemoteCommands),
		)
	}
}

func WrapForHandlers(handler func(p *paho.Publish) error) func(p *paho.Publish) {
	return func(p *paho.Publish) {
		if err := handler(p); err != nil {
			helper.EventHandlerErrorHandler(err)
		}
	}
}
