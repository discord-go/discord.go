package client

import (
	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/rest"
	"github.com/gorilla/websocket"
	"sync"
)

// wsConnection is a default websocket connection wrapper.
type wsConnection struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func (w *wsConnection) Read() ([]byte, error) {
	_, msg, err := w.conn.ReadMessage()
	return msg, err
}

func (w *wsConnection) Write(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func (w *wsConnection) Close() error {
	return w.conn.Close()
}

// Client represents the main Discord client.
type Client struct {
	Rest    *rest.Client
	Gateway *gateway.ShardManager
	Cache   cache.Cache
}

// New creates a new Client with the given token and options.
func New(token string, opts ...Option) *Client {
	cfg := &Config{
		Token: token,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	sm := gateway.NewShardManager(token, 1, cfg.Intents)
	sm.SetConnectionFactory(func(shardID gateway.ShardID) (gateway.Connection, error) {
		conn, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=10&encoding=json", nil)
		if err != nil {
			return nil, err
		}
		return &wsConnection{conn: conn}, nil
	})
	sm.SetCache(cfg.Cache)
	sm.SetGatewayURL("wss://gateway.discord.gg/?v=10&encoding=json")

	c := &Client{
		Rest:    rest.New(token, nil, nil),
		Gateway: sm,
		Cache:   cfg.Cache,
	}

	return c
}
