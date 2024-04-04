package wsapi

import (
	"context"

	"github.com/CE-Thesis-2023/ltd/src/internal/ws"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
)

func GetStandardHandler() ws.MessageHandler {
	return func(ctx context.Context, msg *events.CommandRequest) (*events.CommandResponse, error) {
		return nil, nil
	}
}
