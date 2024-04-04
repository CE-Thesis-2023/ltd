package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	custcon "github.com/CE-Thesis-2023/ltd/src/internal/concurrent"
	"github.com/CE-Thesis-2023/ltd/src/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/events"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type WebSocketClient struct {
	options *WebSocketClientOptions
	conn    *websocket.Conn
	pool    *ants.Pool
	mu      sync.Mutex
}

type WebSocketClientOptions struct {
	configs        *configs.WebSocketFeedConfigs
	deviceId       string
	messageHandler MessageHandler

	poolSize int
}

type WebSocketClientOptioner func(o *WebSocketClientOptions)

func WithGlobalConfigs(c *configs.WebSocketFeedConfigs) WebSocketClientOptioner {
	return func(o *WebSocketClientOptions) {
		o.configs = c
	}
}

type MessageHandler func(ctx context.Context, msg *events.CommandRequest) (*events.CommandResponse, error)

func WithMessageHandler(f MessageHandler) WebSocketClientOptioner {
	return func(o *WebSocketClientOptions) {
		o.messageHandler = f
	}
}

func WithDeviceId(id string) WebSocketClientOptioner {
	return func(o *WebSocketClientOptions) {
		o.deviceId = id
	}
}

func WithPoolSize(size int) WebSocketClientOptioner {
	return func(o *WebSocketClientOptions) {
		o.poolSize = size
	}
}

func NewWebSocketClient(options ...WebSocketClientOptioner) *WebSocketClient {
	opts := &WebSocketClientOptions{}
	for _, option := range options {
		option(opts)
	}
	client := &WebSocketClient{
		options: opts,
		pool:    custcon.New(opts.poolSize),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		logger.SFatal("NewWebSocketClient: Connect fatal", zap.Error(err))
		return nil
	}

	return client
}

func (c *WebSocketClient) Connect(ctx context.Context) error {
	host := c.options.configs.Host
	if c.options.configs.Port != 0 {
		host = fmt.Sprintf("%s:%d", host, c.options.configs.Port)
	}
	path := ""
	if c.options.configs.UpgradePath != "" {
		path = fmt.Sprintf("%s/%s", c.options.configs.UpgradePath, c.options.deviceId)
	}
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path,
	}
	encoded := u.String()
	conn, resp, err := websocket.DefaultDialer.DialContext(
		ctx, encoded, http.Header{})
	if err != nil {
		logger.SError("WebSocketClient.Connect: error", zap.Error(err))
		return err
	}

	if err := c.handleHttpCode(resp.StatusCode); err != nil {
		logger.SError("WebSocketClient.Connect: http code error",
			zap.Error(err))
		return err
	}

	c.conn = conn
	logger.SInfo("WebSocketClient.Connect: connected")
	return nil
}

func (c *WebSocketClient) handleHttpCode(code int) error {
	switch code {
	case 200, 201, 202:
		return nil
	default:
		logger.SDebug("handleHttpCode: code error",
			zap.Int("code", code))
		return custerror.ErrorInternal
	}
}

func (c *WebSocketClient) Run() error {
	logger.SDebug("WebSocketClient.Run: started")
	conn := c.conn
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if errors.Is(err, websocket.ErrCloseSent) {
				logger.SDebug("WebSocketClient.Run: server closed connection")
				return nil
			}
			logger.SError("WebSocketClient.Run: ReadMessage error", zap.Error(err))
			return err
		}
		if c.options.messageHandler != nil {
			c.pool.Submit(func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()

				var msgWithId WebSocketMessageRequest
				if err := json.Unmarshal(msg, &msgWithId); err != nil {
					logger.SError("WebSocketClient.Run: unmarshal error", zap.Error(err))
					return
				}
				resp, err := c.options.messageHandler(ctx, msgWithId.Request)
				if err != nil {
					logger.SError("WebSocketClient.Run: messageHandler error",
						zap.Error(err),
						zap.Uint64("messageId", msgWithId.MessageId))
					return
				}

				respWithId := &WebSocketMessageResponse{
					MessageId: msgWithId.MessageId,
					Response:  resp,
				}

				sendMessage, err := json.Marshal(respWithId)
				if err != nil {
					logger.SError("WebSocketClient.Run: marshal message error",
						zap.Error(err),
						zap.Uint64("messageId", msgWithId.MessageId))
					return
				}
				c.mu.Lock()
				if err := conn.WriteMessage(websocket.TextMessage, sendMessage); err != nil {
					logger.SError("WebSocketClient.Run: WriteMessage error",
						zap.Error(err),
						zap.Uint64("messageId", msgWithId.MessageId))
					return
				}
				c.mu.Unlock()

				logger.SInfo("WebSocketClient.Run: message sent", zap.Uint64("messageId", msgWithId.MessageId))
			})
		}
	}
}

func (c *WebSocketClient) Stop(ctx context.Context) error {
	if err := c.conn.Close(); err != nil {
		return err
	}
	logger.SDebug("WebSocketClient.Stop: shutdown completed")
	return nil
}
