package eventsapi

import (
	"context"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/biz/reconciler"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"

	"encoding/json"

	"github.com/eclipse/paho.golang/paho"
	"go.uber.org/zap"
)

type StandardEventHandler struct {
}

func NewStandardEventHandler() *StandardEventHandler {
	return &StandardEventHandler{}
}

func (h *StandardEventHandler) ReceiveRemoteCommands(p *paho.Publish) error {
	logger.SDebug("received MQTT remote command", zap.String("message", string(p.Payload)))

	var msg events.CommandRequest
	if err := json.Unmarshal(p.Payload, &msg); err != nil {
		logger.SError("failed to parse MQTT remote command",
			zap.Error(err))
		return err
	}

	go func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			time.Second*2)
		defer cancel()
		if err := reconciler.
			GetReconciler().
			ProcessInputEvent(ctx, &msg); err != nil {
			logger.SError("failed to process MQTT remote command",
				zap.Error(err))
			return
		}
		defer func() {
			if ctx.Err() != nil {
				logger.SDebug("remote command processing timeout")
			}
		}()

	}()

	logger.SDebug("process MQTT command success")
	return nil
}
