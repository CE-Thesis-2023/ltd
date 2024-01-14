package wsapi

import (
	"context"

	"github.com/CE-Thesis-2023/ltd/src/biz/handlers"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/internal/ws"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"go.uber.org/zap"
)

func GetStandardHandler() ws.MessageHandler {
	return func(ctx context.Context, msg *events.CommandRequest) (*events.CommandResponse, error) {
		resp, err := handlers.CommandHandlers(ctx, msg)
		if err != nil {
			logger.SDebug("WebSocket.GetStandardHandler: CommandHandlers error", zap.Error(err))
			return nil, err
		}
		logger.SDebug("WebSocket.GetStandardHandler: response", zap.Any("response", resp))
		return resp, nil
	}
}
