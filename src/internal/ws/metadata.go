package ws

import "github.com/CE-Thesis-2023/ltd/src/models/events"

type WebSocketMessageRequest struct {
	MessageId uint64                 `json:"messageId"`
	Request   *events.CommandRequest `json:"commandRequest"`
}

type WebSocketMessageResponse struct {
	MessageId uint64                  `json:"messageId"`
	Response  *events.CommandResponse `json:"commandResponse"`
}
