package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/snowflake"

	"github.com/gorilla/websocket"
)

// Start connects to Discord and begins processing events. It returns after the
// initial connection has been created; use Wait or Done to observe termination.
func (b *Bot) Start(ctx context.Context) error {
	if strings.TrimSpace(b.token) == "" {
		return ErrMissingToken
	}
	if !isValidBotToken(b.token) {
		return ErrInvalidToken
	}
	if ctx == nil {
		ctx = context.Background()
	}

	b.lifecycleMu.Lock()
	defer b.lifecycleMu.Unlock()
	b.stateMu.Lock()
	if b.state == BotStateStarting || b.state == BotStateRunning || b.state == BotStateStopping {
		b.stateMu.Unlock()
		return ErrBotAlreadyRunning
	}
	b.state = BotStateStarting
	b.runDone = nil
	b.runErr = nil
	b.runCtx = nil
	b.runCancel = nil
	b.readyCh = make(chan struct{})
	b.ready = false
	b.readyAt = time.Time{}
	b.user = nil
	b.stateMu.Unlock()

	factory := b.connFactory
	if factory == nil {
		factory = defaultConnectionFactory
	}
	gatewayURL := b.gatewayURL
	if b.compression {
		parsed, err := url.Parse(gatewayURL)
		if err != nil {
			b.setStopped(err)
			return err
		}
		query := parsed.Query()
		query.Set("compress", "zlib-stream")
		parsed.RawQuery = query.Encode()
		gatewayURL = parsed.String()
	}
	connect := func(target string) (gateway.Connection, error) {
		if target == "" {
			target = b.gatewayURL
		}
		if b.compression {
			parsed, err := url.Parse(target)
			if err != nil {
				return nil, err
			}
			query := parsed.Query()
			query.Set("compress", "zlib-stream")
			parsed.RawQuery = query.Encode()
			target = parsed.String()
		}
		return factory(target)
	}
	dispatcher := gateway.NewDispatcher()
	dispatcher.AddHandler(b.handleRawDispatch)
	var client *gateway.Client
	var manager *gateway.ShardManager
	if b.sharded {
		manager = gateway.NewShardManager(b.token, b.shardCount, b.intentsVal)
		manager.Dispatcher = dispatcher
		manager.SetCache(b.cacheStore)
		manager.SetGatewayURL(gatewayURL)
		manager.SetCompression(b.compression)
		manager.SetConnectionURLFactory(func(target string, _ gateway.ShardID) (gateway.Connection, error) {
			return connect(target)
		})
	} else {
		conn, err := connect(gatewayURL)
		if err != nil {
			b.setStopped(err)
			return err
		}
		if conn == nil {
			err = errors.New("bot: connection factory returned a nil connection")
			b.setStopped(err)
			return err
		}
		client = gateway.NewClient(conn, dispatcher)
		client.Session = gateway.NewSession()
		if b.cacheStore != nil {
			client.Cache = b.cacheStore
		}
		client.SetToken(b.token)
		client.Intents = b.intentsVal
		client.GatewayURL = gatewayURL
		client.ConnFactory = connect
	}

	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	b.stateMu.Lock()
	b.gwClient = client
	b.shardManager = manager
	b.runCtx = runCtx
	b.runCancel = cancel
	b.runDone = done
	b.runErr = nil
	b.startedAt = time.Now()
	b.state = BotStateRunning
	b.stateMu.Unlock()

	if manager != nil {
		go b.runShards(manager, runCtx, done)
	} else {
		go b.runGateway(client, runCtx, done)
	}
	return nil
}

// Run starts the bot and blocks until SIGINT, SIGTERM, or a fatal gateway
// error. It performs a graceful shutdown and waits for active handlers.
func (b *Bot) Run() error {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()
	return b.RunContext(ctx)
}

// RunContext starts the bot and runs it until ctx is cancelled or the gateway
// stops. It is the non-signal variant of Run for services and tests.
func (b *Bot) RunContext(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := b.Start(ctx); err != nil {
		return err
	}
	done := b.Done()
	select {
	case <-ctx.Done():
		return b.Stop(context.Background())
	case <-done:
		return b.Wait()
	}
}

// Stop cancels the gateway and waits for active handlers to finish.
func (b *Bot) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	b.lifecycleMu.Lock()
	defer b.lifecycleMu.Unlock()
	b.stateMu.Lock()
	done := b.runDone
	cancel := b.runCancel
	client := b.gwClient
	manager := b.shardManager
	state := b.state
	if state == BotStateRunning || state == BotStateStarting {
		b.state = BotStateStopping
	}
	b.stateMu.Unlock()
	if done == nil || (state != BotStateRunning && state != BotStateStarting && state != BotStateStopping) {
		return nil
	}
	if cancel != nil {
		cancel()
	}
	if client != nil && client.Conn != nil {
		_ = client.Conn.Close()
	}
	if manager != nil {
		if err := manager.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
	b.cancelJobs()
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := b.waitHandlers(ctx); err != nil {
		return err
	}
	if err := b.lastRunError(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// Wait blocks until the current or most recent run finishes.
func (b *Bot) Wait() error {
	done := b.Done()
	if done == nil {
		return ErrBotNotRunning
	}
	<-done
	if err := b.waitHandlers(context.Background()); err != nil {
		return err
	}
	return b.lastRunError()
}

// Done returns a channel closed when the current run terminates.
func (b *Bot) Done() <-chan struct{} {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.runDone
}

// State returns the current lifecycle state.
func (b *Bot) State() BotState {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.state
}

// AppID returns the bot's application/user ID, available after READY.
func (b *Bot) AppID() snowflake.ID {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.appID
}

// Stats returns a snapshot of runtime counters.
func (b *Bot) Stats() BotStats {
	b.stateMu.RLock()
	startedAt := b.startedAt
	b.stateMu.RUnlock()
	return BotStats{
		StartedAt:      startedAt,
		EventsReceived: b.eventsReceived.Load(),
		HandlerPanics:  b.handlerPanics.Load(),
		CommandSyncs:   b.commandSyncs.Load(),
		PrefixCommands: b.prefixCommands.Load(),
		SlashCommands:  b.slashCommands.Load(),
	}
}

func defaultConnectionFactory(url string) (gateway.Connection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &wsConn{conn: conn}, nil
}

type wsConn struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func (w *wsConn) Read() ([]byte, error) {
	_, msg, err := w.conn.ReadMessage()
	return msg, err
}

func (w *wsConn) Write(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func (w *wsConn) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.Close()
}

func (b *Bot) runGateway(client *gateway.Client, ctx context.Context, done chan struct{}) {
	err := client.Start(ctx)
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		b.reportError(fmt.Errorf("gateway: %w", err))
	}
	b.finishRun(err, done)
}

func (b *Bot) runShards(manager *gateway.ShardManager, ctx context.Context, done chan struct{}) {
	err := manager.Start(ctx)
	if err == nil {
		<-ctx.Done()
		err = manager.Shutdown(context.Background())
	}
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		b.reportError(fmt.Errorf("shards: %w", err))
	}
	b.finishRun(err, done)
}

func (b *Bot) finishRun(err error, done chan struct{}) {
	b.cancelJobs()
	b.stateMu.Lock()
	b.runErr = err
	b.state = BotStateStopped
	b.runCancel = nil
	b.stateMu.Unlock()
	b.mu.RLock()
	disconnectHandlers := append([]LifecycleHandler(nil), b.disconnectHandlers...)
	b.mu.RUnlock()
	b.dispatchLifecycle(disconnectHandlers, "DISCONNECT")
	close(done)
}

func (b *Bot) setStopped(err error) {
	b.stateMu.Lock()
	b.state = BotStateStopped
	b.runErr = err
	b.stateMu.Unlock()
}

func (b *Bot) lastRunError() error {
	b.stateMu.RLock()
	defer b.stateMu.RUnlock()
	return b.runErr
}

func (b *Bot) waitHandlers(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		b.handlerWG.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Bot) baseContext() BaseContext {
	b.stateMu.RLock()
	ctx := b.runCtx
	b.stateMu.RUnlock()
	if ctx == nil {
		ctx = context.Background()
	}
	return BaseContext{Bot: b, rest: b.Rest, ctx: ctx, timeout: 10 * time.Second}
}

func (b *Bot) baseContextWithRaw(raw json.RawMessage) BaseContext {
	base := b.baseContext()
	base.raw = append(json.RawMessage(nil), raw...)
	return base
}
