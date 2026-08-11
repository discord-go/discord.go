package bot

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
)

func TestRouterPrefixParsingAliasesAndMiddleware(t *testing.T) {
	router := NewRouter()
	var got []string
	var middlewareRan bool
	command := router.Prefix("say", func(_ *MessageContext, args []string) {
		got = append([]string(nil), args...)
	})
	command.Aliases("s").Use(func(next PrefixHandler) PrefixHandler {
		return func(ctx *MessageContext, args []string) {
			middlewareRan = true
			next(ctx, args)
		}
	})

	ctx := &MessageContext{MessageCreate: &events.MessageCreate{Message: messages.Message{Content: `!s "hello world" plain\ value`}}}
	if !router.handlePrefix(ctx, "!") {
		t.Fatal("expected prefix command to match")
	}
	want := []string{"hello world", "plain value"}
	if !middlewareRan || !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func TestRouterCommandValidationAndStableOrder(t *testing.T) {
	router := NewRouter()
	router.Command("zulu", "Z", func(*InteractionContext) {})
	router.Command("alpha", "A", func(*InteractionContext) {})
	if err := router.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	commands := router.Commands()
	if len(commands) != 2 || commands[0].Name != "alpha" || commands[1].Name != "zulu" {
		t.Fatalf("unexpected command order: %#v", commands)
	}
	if _, err := router.CommandE("Bad Name", "invalid", func(*InteractionContext) {}); err == nil {
		t.Fatal("expected invalid command name error")
	}
	if _, err := router.CommandE("ok", "", func(*InteractionContext) {}); err == nil {
		t.Fatal("expected invalid description error")
	}
}

func TestInteractionOptionsPreserveSnowflakePrecision(t *testing.T) {
	interaction := &interactions.Interaction{
		Type: interactions.InteractionTypeApplicationCommand,
		Data: json.RawMessage(`{"name":"lookup","type":1,"options":[{"name":"target","type":6,"value":9007199254740993},{"name":"nested","type":1,"options":[{"name":"count","type":4,"value":42}]}]}`),
	}
	ctx := newInteractionContext(BaseContext{}, interaction)
	if got, want := ctx.GetUserID("target").String(), "9007199254740993"; got != want {
		t.Fatalf("user option = %s, want %s", got, want)
	}
	if got := ctx.GetIntOption("count"); got != 42 {
		t.Fatalf("nested integer option = %d, want 42", got)
	}
	if got := ctx.Subcommand(); got != "nested" {
		t.Fatalf("subcommand = %q, want nested", got)
	}
}

func TestBotLifecycleWithInjectedConnection(t *testing.T) {
	connection := newBlockingConnection()
	b := New("MTIz.NjQ1.abc123", WithConnectionFactory(func(string) (gateway.Connection, error) {
		return connection, nil
	}), WithCommandSyncDisabled())

	if err := b.Start(nil); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if b.State() != BotStateRunning {
		t.Fatalf("state = %v, want running", b.State())
	}
	if err := b.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if b.State() != BotStateStopped {
		t.Fatalf("state = %v, want stopped", b.State())
	}
}

func TestBotStartRequiresToken(t *testing.T) {
	if err := New("").Start(nil); !errors.Is(err, ErrMissingToken) {
		t.Fatalf("Start() error = %v, want ErrMissingToken", err)
	}
}

func TestBotStartRejectsInvalidTokenFormat(t *testing.T) {
	// A token without three dot-separated segments is not a valid bot token.
	if err := New("not-a-valid-token").Start(nil); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Start() error = %v, want ErrInvalidToken", err)
	}
}

func TestInteractionRoutesAndModalOptions(t *testing.T) {
	router := NewRouter()
	buttonCalled := make(chan string, 1)
	router.Button("confirm:42", func(ctx *InteractionContext) {
		buttonCalled <- ctx.CustomID()
	})
	button := &interactions.Interaction{
		Type: interactions.InteractionTypeMessageComponent,
		Data: json.RawMessage(`{"custom_id":"confirm:42","component_type":2}`),
	}
	if !router.handleInteraction(newInteractionContext(BaseContext{}, button)) {
		t.Fatal("expected button route to match")
	}
	select {
	case got := <-buttonCalled:
		if got != "confirm:42" {
			t.Fatalf("custom ID = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("button handler did not run")
	}

	modal := &interactions.Interaction{
		Type: interactions.InteractionTypeModalSubmit,
		Data: json.RawMessage(`{"custom_id":"profile","components":[{"type":1,"components":[{"type":4,"custom_id":"name","value":"Ada"}]}]}`),
	}
	modalContext := newInteractionContext(BaseContext{}, modal)
	if got := modalContext.ModalValue("name"); got != "Ada" {
		t.Fatalf("modal value = %q, want Ada", got)
	}
	if !modalContext.IsModalSubmit() {
		t.Fatal("expected modal submit")
	}

	// Test ModalRows for structured access.
	rows := modalContext.ModalRows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0].Components) != 1 {
		t.Fatalf("expected 1 field in row 0, got %d", len(rows[0].Components))
	}
	if rows[0].Components[0].CustomID != "name" || rows[0].Components[0].Value != "Ada" {
		t.Errorf("row 0 field = %+v, want {name Ada}", rows[0].Components[0])
	}
}

func TestGenericEventSubscriptionAndCacheFacade(t *testing.T) {
	b := New("token", WithCache(cache.NewMemoryCache()))
	guildID := snowflake.ID(123)
	store := b.cacheStore.(*cache.MemoryCache)
	store.SetGuild(guildID.String(), &guilds.Guild{ID: guildID, Name: "Test"})
	guild, ok := b.CachedGuild(guildID)
	if !ok || guild.Name != "Test" {
		t.Fatalf("cached guild = %#v, %v", guild, ok)
	}

	called := make(chan string, 1)
	b.OnceEvent("CUSTOM_EVENT", func(ctx *EventContext) { called <- ctx.Name })
	b.handleRawDispatch([]byte(`{"op":0,"t":"CUSTOM_EVENT","d":{"ok":true}}`))
	select {
	case got := <-called:
		if got != "CUSTOM_EVENT" {
			t.Fatalf("event name = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("generic event handler did not run")
	}
	if b.Stats().EventsReceived != 1 {
		t.Fatalf("events received = %d, want 1", b.Stats().EventsReceived)
	}
}

func TestReconnectLifecycleHandler(t *testing.T) {
	b := New("token")
	called := make(chan struct{}, 1)
	b.OnReconnect(func() { called <- struct{}{} })
	b.handleRawDispatch([]byte(`{"op":7,"d":null}`))
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("reconnect handler did not run")
	}
}

func TestTemplateMessageTriggers(t *testing.T) {
	b := New("token", WithPrefix("?"), WithBotName("zar"), WithMentionTriggers(true))
	b.stateMu.Lock()
	b.appID = snowflake.ID(123)
	b.stateMu.Unlock()
	for _, test := range []struct {
		content string
		want    string
	}{
		{content: "?ping", want: "ping"},
		{content: "zar ping", want: "ping"},
		{content: "<@123> ping", want: "ping"},
		{content: "<@!123> ping", want: "ping"},
	} {
		ctx := &MessageContext{MessageCreate: &events.MessageCreate{Message: messages.Message{Content: test.content}}}
		if got := b.messageCommandContent(ctx); got != test.want {
			t.Fatalf("trigger %q content = %q, want %q", test.content, got, test.want)
		}
	}
}

func TestInteractionCollector(t *testing.T) {
	b := New("token")
	result := make(chan *InteractionContext, 1)
	go func() {
		ctx, err := b.AwaitInteraction(context.Background(), func(ctx *InteractionContext) bool {
			return ctx.CustomID() == "collect-me"
		})
		if err == nil {
			result <- ctx
		}
	}()
	time.Sleep(time.Millisecond)
	b.handleRawDispatch([]byte(`{"op":0,"t":"INTERACTION_CREATE","d":{"id":"1","type":3,"custom_id":"collect-me","data":{"custom_id":"collect-me","component_type":2}}}`))
	select {
	case ctx := <-result:
		if ctx.CustomID() != "collect-me" {
			t.Fatalf("collected custom ID = %q", ctx.CustomID())
		}
	case <-time.After(time.Second):
		t.Fatal("interaction was not collected")
	}
}

type blockingConnection struct {
	closed chan struct{}
}

func newBlockingConnection() *blockingConnection {
	return &blockingConnection{closed: make(chan struct{})}
}

func (c *blockingConnection) Read() ([]byte, error) {
	<-c.closed
	return nil, errors.New("connection closed")
}

func (c *blockingConnection) Write([]byte) error { return nil }

func (c *blockingConnection) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return nil
}

func TestCollectInteractions(t *testing.T) {
	b := New("token")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, stop := b.CollectInteractions(ctx, func(ic *InteractionContext) bool {
		return ic.CustomID() == "vote:yes"
	})
	defer stop()

	// Dispatch two matching interactions.
	b.handleRawDispatch([]byte(`{"op":0,"t":"INTERACTION_CREATE","d":{"id":"1","type":3,"data":{"custom_id":"vote:yes","component_type":2}}}`))
	b.handleRawDispatch([]byte(`{"op":0,"t":"INTERACTION_CREATE","d":{"id":"2","type":3,"data":{"custom_id":"vote:yes","component_type":2}}}`))

	// Should receive both.
	count := 0
	for {
		select {
		case ic := <-ch:
			count++
			_ = ic
			if count >= 2 {
				return
			}
		case <-time.After(time.Second):
			t.Fatalf("expected 2 interactions, got %d", count)
		}
	}
}
