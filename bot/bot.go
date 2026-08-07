package bot

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/rest"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/storage"
	"github.com/discord-go/discord.go/users"
)

// ReadyHandler is called when the bot successfully connects to Discord.
type ReadyHandler func(ctx *ReadyContext)

// MessageHandler is called on every MESSAGE_CREATE event. Bot-authored
// messages are filtered before this handler is called.
type MessageHandler func(ctx *MessageContext)

// MessageUpdateHandler is called on MESSAGE_UPDATE events.
type MessageUpdateHandler func(ctx *MessageUpdateContext)

// MessageDeleteHandler is called on MESSAGE_DELETE events.
type MessageDeleteHandler func(ctx *MessageDeleteContext)

// InteractionHandler is called on every INTERACTION_CREATE event.
type InteractionHandler func(ctx *InteractionContext)

// ReactionHandler is called when a reaction is added to a message.
type ReactionHandler func(ctx *ReactionContext)

// GuildCreateHandler is called when the bot joins a guild or on initial READY.
type GuildCreateHandler func(ctx *GuildContext)

// GuildUpdateHandler is called on GUILD_UPDATE events.
type GuildUpdateHandler func(ctx *GuildUpdateContext)

// GuildDeleteHandler is called on GUILD_DELETE events.
type GuildDeleteHandler func(ctx *GuildDeleteContext)

// ChannelCreateHandler is called on CHANNEL_CREATE events.
type ChannelCreateHandler func(ctx *ChannelContext)

// ChannelUpdateHandler is called on CHANNEL_UPDATE events.
type ChannelUpdateHandler func(ctx *ChannelUpdateContext)

// GuildAuditLogEntryCreateHandler is called on GUILD_AUDIT_LOG_ENTRY_CREATE events.
type GuildAuditLogEntryCreateHandler func(ctx *GuildAuditLogEntryContext)

// RawEventHandler receives every dispatch event, including events that do not
// yet have a typed helper in this package.
type RawEventHandler func(ctx context.Context, event string, data json.RawMessage)

// EventHandler is the catch-all typed event handler. It gives applications
// access to every Discord dispatch, including newer events not yet modeled by
// this module.
type EventHandler func(ctx *EventContext)

// LifecycleHandler is called when the gateway lifecycle changes.
type LifecycleHandler func()

// ErrorHandler receives gateway, parsing, synchronization, and handler panic
// errors. The default handler logs the error.
type ErrorHandler func(error)

// ConnectionFactory creates a gateway connection. It is primarily useful for
// tests, proxies, and applications that provide their own WebSocket transport.
type ConnectionFactory func(url string) (gateway.Connection, error)

// BotState describes the bot lifecycle state.
type BotState uint8

const (
	BotStateStopped BotState = iota
	BotStateStarting
	BotStateRunning
	BotStateStopping
)

// CommandSyncMode controls where router commands are registered.
type CommandSyncMode uint8

const (
	// CommandSyncGlobal registers commands globally. Global changes can take
	// up to an hour to propagate in Discord.
	CommandSyncGlobal CommandSyncMode = iota
	// CommandSyncGuild registers commands in one guild and is useful during
	// development because changes propagate almost immediately.
	CommandSyncGuild
	// CommandSyncDisabled leaves command registration to the application.
	CommandSyncDisabled
)

// CommandSyncConfig configures automatic command registration.
type CommandSyncConfig struct {
	Mode    CommandSyncMode
	GuildID snowflake.ID
	Timeout time.Duration
}

// BotStats is a point-in-time snapshot of runtime counters.
type BotStats struct {
	StartedAt      time.Time
	EventsReceived uint64
	HandlerPanics  uint64
	CommandSyncs   uint64
	PrefixCommands uint64
	SlashCommands  uint64
}

// Bot is the main entry point for building Discord bots with discord.go.
// It wraps the gateway connection, REST client, and event dispatching into a
// single object with explicit lifecycle control.
type Bot struct {
	mu sync.RWMutex

	token           string
	intentsVal      intents.Intent
	prefix          string
	botName         string
	mentionTriggers bool
	compression     bool
	sharded         bool
	shardCount      int

	// Rest is the Discord REST API client. You can use it directly for API
	// calls not covered by context convenience methods.
	Rest *rest.Client

	// Store is an optional persistence backend for application state.
	Store storage.Store

	cacheStore cache.Cache

	gwClient     *gateway.Client
	shardManager *gateway.ShardManager
	router       *Router
	appID        snowflake.ID
	gatewayURL   string
	connFactory  ConnectionFactory
	logger       *log.Logger
	errorHandler ErrorHandler

	readyHandlers       []ReadyHandler
	messageHandlers     []MessageHandler
	messageUpdate       []MessageUpdateHandler
	messageDelete       []MessageDeleteHandler
	interactionHandlers []InteractionHandler
	reactionHandlers    []ReactionHandler
	guildCreateHandlers []GuildCreateHandler
	guildUpdateHandlers []GuildUpdateHandler
	guildDeleteHandlers []GuildDeleteHandler
	channelCreate       []ChannelCreateHandler
	channelUpdate       []ChannelUpdateHandler
	auditLogCreate      []GuildAuditLogEntryCreateHandler
	rawHandlers         []RawEventHandler
	errorHandlers       []ErrorHandler
	eventHandlers       map[string][]eventSubscription
	reconnectHandlers   []LifecycleHandler
	resumeHandlers      []LifecycleHandler
	invalidatedHandlers []LifecycleHandler
	disconnectHandlers  []LifecycleHandler

	commandSync CommandSyncConfig
	syncMu      sync.Mutex
	presenceMu  sync.Mutex
	presence    *PresenceUpdate

	lifecycleMu sync.Mutex
	stateMu     sync.RWMutex
	state       BotState
	runCtx      context.Context
	runCancel   context.CancelFunc
	runDone     chan struct{}
	runErr      error
	startedAt   time.Time
	readyCh     chan struct{}
	ready       bool
	readyAt     time.Time
	user        *users.User

	handlerWG             sync.WaitGroup
	handlerSlots          chan struct{}
	collectorMu           sync.Mutex
	interactionCollectors map[uint64]interactionCollector
	messageCollectors     map[uint64]messageCollector
	reactionCollectors    map[uint64]reactionCollector
	jobsMu                sync.Mutex
	jobs                  map[uint64]context.CancelFunc

	eventsReceived atomic.Uint64
	handlerPanics  atomic.Uint64
	commandSyncs   atomic.Uint64
	prefixCommands atomic.Uint64
	slashCommands  atomic.Uint64
	subscriptions  atomic.Uint64
}

// Option configures the Bot.
type Option func(*Bot)

// WithIntents sets the gateway intents for the bot.
// Defaults to Guilds | GuildMessages | MessageContent.
func WithIntents(i intents.Intent) Option {
	return func(b *Bot) { b.intentsVal = i }
}

// WithPrefix sets the prefix for text commands. An empty prefix disables
// prefix command dispatching. The default is "!".
func WithPrefix(prefix string) Option {
	return func(b *Bot) { b.prefix = prefix }
}

// WithBotName enables bot-name message triggers such as "zar ping".
func WithBotName(name string) Option {
	return func(b *Bot) { b.botName = strings.TrimSpace(name) }
}

// WithMentionTriggers enables <@bot-id> and <@!bot-id> message triggers.
func WithMentionTriggers(enabled bool) Option {
	return func(b *Bot) { b.mentionTriggers = enabled }
}

// WithGatewayCompression enables Discord's zlib-stream gateway compression.
func WithGatewayCompression(enabled bool) Option {
	return func(b *Bot) { b.compression = enabled }
}

// WithPresence configures a presence that is sent after READY and reapplied
// after a fresh identify.
func WithPresence(presence PresenceUpdate) Option {
	return func(b *Bot) {
		presence.Activities = append([]Activity(nil), presence.Activities...)
		b.presence = &presence
	}
}

// WithShards enables high-level gateway sharding. A count of zero asks Discord
// for the recommended shard count.
func WithShards(count int) Option {
	return func(b *Bot) {
		b.sharded = true
		if count > 0 {
			b.shardCount = count
		}
	}
}

// WithRouter attaches a command router for slash and prefix commands.
func WithRouter(r *Router) Option {
	return func(b *Bot) { b.router = r }
}

// WithRESTClient replaces the default REST client. This is useful for custom
// HTTP transports, test servers, and custom rate-limit stores.
func WithRESTClient(client *rest.Client) Option {
	return func(b *Bot) { b.Rest = client }
}

// WithCache enables typed cache lookups and gateway cache hydration.
func WithCache(store cache.Cache) Option {
	return func(b *Bot) { b.cacheStore = store }
}

// WithStore attaches a persistence backend for application-owned data.
func WithStore(store storage.Store) Option {
	return func(b *Bot) { b.Store = store }
}

// WithLogger sets the logger used for lifecycle and runtime diagnostics. A
// nil logger restores the standard logger.
func WithLogger(logger *log.Logger) Option {
	return func(b *Bot) {
		if logger == nil {
			b.logger = log.Default()
			return
		}
		b.logger = logger
	}
}

// WithErrorHandler sets the default error callback. Additional callbacks can
// be registered with OnError.
func WithErrorHandler(handler ErrorHandler) Option {
	return func(b *Bot) {
		if handler != nil {
			b.errorHandler = handler
		}
	}
}

// WithGatewayURL overrides the Discord gateway URL. It is useful when using a
// gateway proxy or a local test server.
func WithGatewayURL(url string) Option {
	return func(b *Bot) {
		if url != "" {
			b.gatewayURL = url
		}
	}
}

// WithConnectionFactory supplies a custom gateway connection factory.
func WithConnectionFactory(factory ConnectionFactory) Option {
	return func(b *Bot) {
		if factory != nil {
			b.connFactory = factory
		}
	}
}

// WithMaxHandlerConcurrency limits the number of event handlers running at
// once. Zero leaves concurrency unlimited.
func WithMaxHandlerConcurrency(limit int) Option {
	return func(b *Bot) {
		if limit > 0 {
			b.handlerSlots = make(chan struct{}, limit)
		}
	}
}

// WithCommandSync configures automatic router command registration on READY.
// Automatic synchronization remains enabled globally by default when a
// router is attached.
func WithCommandSync(config CommandSyncConfig) Option {
	return func(b *Bot) {
		if config.Timeout <= 0 {
			config.Timeout = 30 * time.Second
		}
		b.commandSync = config
	}
}

// WithCommandSyncDisabled disables automatic command registration.
func WithCommandSyncDisabled() Option {
	return WithCommandSync(CommandSyncConfig{Mode: CommandSyncDisabled})
}

// WithGuildCommandSync enables fast, guild-scoped command synchronization.
func WithGuildCommandSync(guildID snowflake.ID) Option {
	return WithCommandSync(CommandSyncConfig{Mode: CommandSyncGuild, GuildID: guildID})
}

// New creates a new Bot instance.
func New(token string, opts ...Option) *Bot {
	b := &Bot{
		token:      token,
		intentsVal: intents.Guilds | intents.GuildMessages | intents.MessageContent,
		prefix:     "!",
		gatewayURL: "wss://gateway.discord.gg/?v=10&encoding=json",
		logger:     log.Default(),
		commandSync: CommandSyncConfig{
			Mode:    CommandSyncGlobal,
			Timeout: 30 * time.Second,
		},
		eventHandlers:         make(map[string][]eventSubscription),
		readyCh:               make(chan struct{}),
		interactionCollectors: make(map[uint64]interactionCollector),
		messageCollectors:     make(map[uint64]messageCollector),
		jobs:                  make(map[uint64]context.CancelFunc),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}
	if b.Rest == nil {
		b.Rest = rest.New(token, nil, nil)
	}
	if b.errorHandler == nil {
		b.errorHandler = func(err error) { b.logger.Printf("bot: %s", b.redactToken(err.Error())) }
	}
	if b.router != nil {
		b.router.setErrorHandler(b.reportError)
	}
	return b
}

// OnReady registers a handler for the READY event.
func (b *Bot) OnReady(handler ReadyHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.readyHandlers = append(b.readyHandlers, handler)
	b.mu.Unlock()
}

// OnMessageCreate registers a handler for MESSAGE_CREATE events.
func (b *Bot) OnMessageCreate(handler MessageHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.messageHandlers = append(b.messageHandlers, handler)
	b.mu.Unlock()
}

// OnMessageUpdate registers a handler for MESSAGE_UPDATE events.
func (b *Bot) OnMessageUpdate(handler MessageUpdateHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.messageUpdate = append(b.messageUpdate, handler)
	b.mu.Unlock()
}

// OnMessageDelete registers a handler for MESSAGE_DELETE events.
func (b *Bot) OnMessageDelete(handler MessageDeleteHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.messageDelete = append(b.messageDelete, handler)
	b.mu.Unlock()
}

// OnInteraction registers a handler for all INTERACTION_CREATE events.
func (b *Bot) OnInteraction(handler InteractionHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.interactionHandlers = append(b.interactionHandlers, handler)
	b.mu.Unlock()
}

// OnMessageReactionAdd registers a handler for MESSAGE_REACTION_ADD events.
func (b *Bot) OnMessageReactionAdd(handler ReactionHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.reactionHandlers = append(b.reactionHandlers, handler)
	b.mu.Unlock()
}

// OnGuildCreate registers a handler for GUILD_CREATE events.
func (b *Bot) OnGuildCreate(handler GuildCreateHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.guildCreateHandlers = append(b.guildCreateHandlers, handler)
	b.mu.Unlock()
}

// OnGuildUpdate registers a handler for GUILD_UPDATE events.
func (b *Bot) OnGuildUpdate(handler GuildUpdateHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.guildUpdateHandlers = append(b.guildUpdateHandlers, handler)
	b.mu.Unlock()
}

// OnGuildDelete registers a handler for GUILD_DELETE events.
func (b *Bot) OnGuildDelete(handler GuildDeleteHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.guildDeleteHandlers = append(b.guildDeleteHandlers, handler)
	b.mu.Unlock()
}

// OnRawEvent registers a handler for every gateway dispatch event.
func (b *Bot) OnRawEvent(handler RawEventHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.rawHandlers = append(b.rawHandlers, handler)
	b.mu.Unlock()
}

// OnError registers an additional runtime error handler.
func (b *Bot) OnError(handler ErrorHandler) {
	if handler == nil {
		return
	}
	b.mu.Lock()
	b.errorHandlers = append(b.errorHandlers, handler)
	b.mu.Unlock()
}
