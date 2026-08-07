# Testing Guide

## Overview

discord.go is designed for testability. The gateway connection, REST HTTP
client, and command handlers can all be mocked or replaced in tests. This guide
covers the patterns for testing bots built with the library.

## Mocking the Gateway

The gateway `Connection` interface has three methods:

```go
type Connection interface {
    Read() ([]byte, error)
    Write([]byte) error
    Close() error
}
```

Implement this in test code to simulate gateway events:

```go
type testConnection struct {
    messages [][]byte
    idx      int
    writes   [][]byte
}

func (t *testConnection) Read() ([]byte, error) {
    if t.idx >= len(t.messages) {
        return nil, errors.New("connection closed")
    }
    msg := t.messages[t.idx]
    t.idx++
    return msg, nil
}

func (t *testConnection) Write(data []byte) error {
    t.writes = append(t.writes, data)
    return nil
}

func (t *testConnection) Close() error { return nil }
```

Inject it via `bot.WithConnectionFactory`:

```go
conn := &testConnection{
    messages: [][]byte{
        // Simulate a READY event
        []byte(`{"op":0,"t":"READY","d":{"v":10,"user":{"id":"100","username":"testbot"}}}`),
    },
}

bot := bot.New("test-token",
    bot.WithIntents(intents.Guilds),
    bot.WithConnectionFactory(func(url string) (gateway.Connection, error) {
        return conn, nil
    }),
)
```

## Mocking REST

Pass a custom `http.Client` to `rest.New` that returns canned responses:

```go
import "net/http/httptest"

func newMockREST(t *testing.T, status int, response string) *rest.Client {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        w.Write([]byte(response))
    }))
    t.Cleanup(server.Close)

    httpClient := &http.Client{Transport: &mockTransport{server: server}}
    return rest.New("test-token", nil, httpClient)
}
```

Or use `httptest.Server` directly to test REST call behavior:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Assert request method, path, headers, body
    if r.URL.Path != "/channels/123/messages" {
        t.Errorf("unexpected path: %s", r.URL.Path)
    }
    w.Write([]byte(`{"id":"456","content":"hello"}`))
}))
defer server.Close()
```

## Testing Command Handlers

Command handlers receive an `*InteractionContext`. Create one manually in tests:

```go
func TestPingCommand(t *testing.T) {
    router := bot.NewRouter()
    called := false
    router.Command("ping", "Check bot status", func(ctx *bot.InteractionContext) {
        called = true
        ctx.Reply("Pong!")
    })

    // Build a mock interaction
    interaction := interactions.Interaction{
        ID:   "100",
        Type: interactions.InteractionTypeApplicationCommand,
        Data: json.RawMessage(`{"name":"ping"}`),
    }

    // Create context and invoke router
    ctx := bot.NewTestInteractionContext(&interaction)
    router.handleInteraction(ctx)

    if !called {
        t.Error("handler was not called")
    }
}
```

## Testing Collectors

Use `bot.AwaitInteraction` with a context timeout in tests:

```go
func TestButtonCollector(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    result, err := botClient.AwaitInteraction(ctx, func(ic *bot.InteractionContext) bool {
        return ic.CustomID() == "confirm"
    })

    if err != nil {
        t.Fatalf("await interaction: %v", err)
    }
    // Assert on result
}
```

`bot.AwaitReaction` works the same way for reaction collectors:

```go
func TestReactionCollector(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    result, err := botClient.AwaitReaction(ctx, func(rc *bot.ReactionContext) bool {
        return rc.Emoji.Name == "✅"
    })

    if err != nil {
        t.Fatalf("await reaction: %v", err)
    }
    // Assert on result
}
```

## Testing Interaction Servers

`interactions.Server` is an `http.Handler` and can be tested with
`httptest.NewRecorder`:

```go
func TestInteractionServer(t *testing.T) {
    publicKey, privateKey, _ := ed25519.GenerateKey(nil)
    publicKeyHex := hex.EncodeToString(publicKey)

    srv := interactions.NewServer(publicKeyHex, func(i *interactions.Interaction) *interactions.InteractionResponse {
        return &interactions.InteractionResponse{
            Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
            Data: &interactions.InteractionCallbackData{Content: "Hello!"},
        }
    })
    // Override time for deterministic replay tests:
    srv.now = func() time.Time { return time.Unix(1700000000, 0) }

    body := []byte(`{"type":2,"token":"test","version":1}`)
    timestamp := "1700000000"
    message := append([]byte(timestamp), body...)
    signature := ed25519.Sign(privateKey, message)

    req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
    req.Header.Set("X-Signature-Timestamp", timestamp)
    req.Header.Set("X-Signature-Ed25519", hex.EncodeToString(signature))
    w := httptest.NewRecorder()
    srv.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
}
```

## Common Patterns

- Use `bot.WithConnectionFactory` to inject test connections.
- Use `httptest.Server` to mock REST endpoints.
- Create `InteractionContext` manually for unit testing handlers.
- Use context timeouts for collector tests.

## Best Practices

- Test the happy path and error cases.
- Assert on the HTTP method, path, headers, and body for REST calls.
- Test middleware in isolation by wrapping a test handler.
- Clean up test servers with `t.Cleanup`.

## Common Mistakes

- Not closing mock servers, causing port leaks.
- Using real tokens in tests. Always use test tokens.
- Not setting timeouts on collector tests, causing hangs.
- Testing the library's internals instead of your bot's behavior.
