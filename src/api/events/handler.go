package eventsapi

import (
	"context"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/biz/handlers"
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/cache"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"

	"github.com/CE-Thesis-2023/ltd/src/internal/concurrent"

	"github.com/bytedance/sonic"
	"github.com/dgraph-io/ristretto"
	"github.com/eclipse/paho.golang/paho"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type StandardEventHandler struct {
	pool  *ants.Pool
	cache *ristretto.Cache
}

func NewStandardEventHandler() *StandardEventHandler {
	return &StandardEventHandler{
		pool:  custcon.New(100),
		cache: cache.Cache(),
	}
}

func (h *StandardEventHandler) ReceiveRemoteCommands(p *paho.Publish) error {
	logger.SDebug("ReceiveRemoteCommands", zap.String("message", string(p.Payload)))

	var msg events.CommandRequest
	if err := sonic.Unmarshal(p.Payload, &msg); err != nil {
		logger.SError("ReceiveRemoteCommands: message parsing failed", zap.Error(err))
		return err
	}

	h.pool.Submit(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer func() {
			if ctx.Err() != nil {
				logger.SDebug("ReceiveRemoteCommands: context exceeded")
			}
			cancel()
		}()

		resp, err := handlers.CommandHandlers(ctx, &msg)
		if err != nil {
			logger.SError("ReceiveRemoteCommands: CommandHandlers error", zap.Error(err))
			return
		}

		if resp != nil {
			logger.SDebug("ReceiveRemoteCommands: CommandHandlers response", zap.Any("response", resp))
			return
		}
	})

	logger.SDebug("ReceiveRemoteCommands: goroutine assigned")
	return nil
}

func (h *StandardEventHandler) ReceiveRemoteMovementControl(p *paho.Publish) error {
	logger.SInfo("ReceiveRemoteMovementControl: received request")
	h.pool.Submit(func() {
		logger.SDebug("ReceiveRemoteMovementControl", zap.String("message", string(p.Payload)))

		var msg events.PtzCtrlRequest
		if err := sonic.Unmarshal(p.Payload, &msg); err != nil {
			logger.SError("ReceiveRemoteMovementControl: message parsing failed", zap.Error(err))
			return
		}

		dur := time.Second * (time.Duration(*msg.StopAfterSeconds) + 2)
		ctx, cancel := context.WithTimeout(context.Background(), dur)
		defer func() {
			if ctx.Err() != nil {
				logger.SDebug("ReceiveRemoteMovementControl: context exceeded")
			}
			cancel()
		}()
		if err := service.GetCommandService().PtzCtrl(ctx, &msg); err != nil {
			logger.SError("ReceiveRemoteMovementControl: CommandService.PtzCtrl", zap.Error(err))
			return
		}

		logger.SDebug("ReceiveRemoteMovementControl: success")
	})
	return nil
}
