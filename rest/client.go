package rest

import (
	"sync"
	"time"

	"github.com/discord-go/discord.go/http"
	"github.com/discord-go/discord.go/ratelimit"
)

// Client is a Discord REST API client.
type Client struct {
	Token      string
	AuthMode   AuthMode
	HTTPClient http.Client
	Limiter    ratelimit.Limiter
	BaseURL    string

	invalidMu         sync.Mutex
	invalidTimestamps [9500]time.Time
	invalidIdx        int
}

// AuthMode controls the authorization header used by REST requests.
type AuthMode string

const (
	AuthBot    AuthMode = "Bot"
	AuthBearer AuthMode = "Bearer"
	AuthNone   AuthMode = ""
)

// New creates a new REST client.
func New(token string, limiter ratelimit.Limiter, httpClient http.Client) *Client {
	if httpClient == nil {
		httpClient = http.NewClient("DiscordBot (https://github.com/discord-go, 1.0)")
	}
	if limiter == nil {
		limiter = ratelimit.NewLimiter(ratelimit.NewMemoryStore())
	}
	return &Client{
		Token:      token,
		AuthMode:   AuthBot,
		HTTPClient: httpClient,
		Limiter:    limiter,
		BaseURL:    "https://discord.com/api/v10",
	}
}

// SetBearerToken configures OAuth2 bearer authentication.
func (c *Client) SetBearerToken(token string) {
	c.Token = token
	c.AuthMode = AuthBearer
}

// SetBotToken configures bot-token authentication.
func (c *Client) SetBotToken(token string) {
	c.Token = token
	c.AuthMode = AuthBot
}
