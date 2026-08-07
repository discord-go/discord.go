// Package oauth2 provides typed helpers for Discord OAuth2 flows.
package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/discord-go/discord.go/application"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/users"
)

const (
	AuthorizeURL = "https://discord.com/oauth2/authorize"
	TokenURL     = "https://discord.com/api/oauth2/token"
	RevokeURL    = "https://discord.com/api/oauth2/token/revoke"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	HTTPClient   *http.Client
}

type Client struct{ Config }

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type AuthorizationInfo struct {
	Scopes      []string `json:"scopes"`
	Expires     string   `json:"expires"`
	Application any      `json:"application"`
}

func New(config Config) *Client {
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return &Client{Config: config}
}

// GenerateState generates a cryptographically random state string for use
// in OAuth2 CSRF protection. The state should be stored in the user's session
// before redirecting to Discord, and verified when Discord redirects back.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("oauth2: failed to generate state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (c *Client) AuthorizationURL(scopes []string, state string) string {
	values := url.Values{}
	values.Set("client_id", c.ClientID)
	values.Set("response_type", "code")
	values.Set("scope", strings.Join(scopes, " "))
	if c.RedirectURI != "" {
		values.Set("redirect_uri", c.RedirectURI)
	}
	if state != "" {
		values.Set("state", state)
	}
	return AuthorizeURL + "?" + values.Encode()
}

func (c *Client) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	return c.tokenRequest(ctx, url.Values{"grant_type": {"authorization_code"}, "code": {code}})
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	return c.tokenRequest(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refreshToken}})
}

func (c *Client) tokenRequest(ctx context.Context, values url.Values) (*TokenResponse, error) {
	values.Set("client_id", c.ClientID)
	values.Set("client_secret", c.ClientSecret)
	if c.RedirectURI != "" {
		values.Set("redirect_uri", c.RedirectURI)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("oauth2: token request failed: %s", strings.TrimSpace(string(body)))
	}
	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (c *Client) RevokeToken(ctx context.Context, token string) error {
	values := url.Values{"client_id": {c.ClientID}, "client_secret": {c.ClientSecret}, "token": {token}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, RevokeURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("oauth2: revoke failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) CurrentUser(ctx context.Context, accessToken string) (*users.User, error) {
	var user users.User
	err := c.bearerRequest(ctx, accessToken, "GET", "https://discord.com/api/v10/users/@me", nil, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) CurrentUserGuilds(ctx context.Context, accessToken string) ([]guilds.Guild, error) {
	var result []guilds.Guild
	err := c.bearerRequest(ctx, accessToken, "GET", "https://discord.com/api/v10/users/@me/guilds", nil, &result)
	return result, err
}

func (c *Client) CurrentApplication(ctx context.Context, accessToken string) (*application.Application, error) {
	var result application.Application
	err := c.bearerRequest(ctx, accessToken, "GET", "https://discord.com/api/v10/oauth2/applications/@me", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) CurrentAuthorization(ctx context.Context, accessToken string) (*AuthorizationInfo, error) {
	var result AuthorizationInfo
	err := c.bearerRequest(ctx, accessToken, "GET", "https://discord.com/api/v10/oauth2/@me", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) bearerRequest(ctx context.Context, accessToken, method, endpoint string, body io.Reader, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("oauth2: request failed: %s", strings.TrimSpace(string(data)))
	}
	if target != nil && len(data) > 0 {
		return json.Unmarshal(data, target)
	}
	return nil
}
