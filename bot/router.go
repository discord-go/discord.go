package bot

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/rest"
)

// CommandHandler handles a slash command interaction.
type CommandHandler func(ctx *InteractionContext)

// PrefixHandler handles a prefix-based text command. Arguments preserve quoted
// values, so `!say "hello world"` produces one argument.
type PrefixHandler func(ctx *MessageContext, args []string)

// Middleware wraps a CommandHandler to add pre/post processing.
type Middleware func(next CommandHandler) CommandHandler

// PrefixMiddleware wraps a PrefixHandler to add pre/post processing.
type PrefixMiddleware func(next PrefixHandler) PrefixHandler

// PrefixValidation validates a prefix command invocation.
type PrefixValidation func(*MessageContext, []string) error

// InteractionRoute handles buttons, select menus, modals, and autocomplete.
type InteractionRoute struct {
	ID      string
	Handler InteractionHandler

	mu         sync.RWMutex
	middleware []Middleware
}

// Use adds middleware to an interaction route.
func (r *InteractionRoute) Use(mw Middleware) *InteractionRoute {
	if r == nil || mw == nil {
		return r
	}
	r.mu.Lock()
	r.middleware = append(r.middleware, mw)
	r.mu.Unlock()
	return r
}

// Command represents a registered slash command with its handler and middleware.
type Command struct {
	Name        string
	Description string
	Type        interactions.ApplicationCommandType
	Options     []interactions.ApplicationCommandOption
	Handler     CommandHandler
	Category    string

	mu         sync.RWMutex
	middleware []Middleware
}

// Use adds middleware to this specific command. Middleware is applied in
// registration order, with the first middleware running outermost.
func (c *Command) Use(mw Middleware) *Command {
	if c == nil || mw == nil {
		return c
	}
	c.mu.Lock()
	c.middleware = append(c.middleware, mw)
	c.mu.Unlock()
	return c
}

// InCategory labels a command for help menus and command loaders.
func (c *Command) InCategory(category string) *Command {
	if c != nil {
		c.Category = strings.TrimSpace(category)
	}
	return c
}

// Cooldown adds a per-user interaction cooldown.
func (c *Command) Cooldown(duration time.Duration) *Command {
	return c.Use(Cooldown(duration))
}

// RequirePermissions adds an all-permissions guard to the command.
func (c *Command) RequirePermissions(perms permissions.Permission) *Command {
	return c.Use(RequirePermissions(perms))
}

// RequireBotPermissions adds a bot-permissions guard to the command.
func (c *Command) RequireBotPermissions(perms permissions.Permission) *Command {
	return c.Use(RequireBotPermissions(perms))
}

// PrefixCommand represents a prefix-based text command.
type PrefixCommand struct {
	Name        string
	Handler     PrefixHandler
	description string
	usage       string
	minArgs     int
	validations []PrefixValidation

	mu         sync.RWMutex
	middleware []PrefixMiddleware
	router     *Router
}

// Use adds middleware to a prefix command.
func (c *PrefixCommand) Use(mw PrefixMiddleware) *PrefixCommand {
	if c == nil || mw == nil {
		return c
	}
	c.mu.Lock()
	c.middleware = append(c.middleware, mw)
	c.mu.Unlock()
	return c
}

// Description sets help text for the command.
func (c *PrefixCommand) Description(description string) *PrefixCommand {
	if c != nil {
		c.description = description
	}
	return c
}

// Usage sets a usage hint returned when the minimum argument count is not met.
func (c *PrefixCommand) Usage(usage string) *PrefixCommand {
	if c != nil {
		c.usage = usage
	}
	return c
}

// MinArgs requires at least count arguments.
func (c *PrefixCommand) MinArgs(count int) *PrefixCommand {
	if c != nil && count > 0 {
		c.minArgs = count
	}
	return c
}

// Validate adds application validation to a prefix command.
func (c *PrefixCommand) Validate(check PrefixValidation) *PrefixCommand {
	if c != nil && check != nil {
		c.validations = append(c.validations, check)
	}
	return c
}

// Aliases adds additional names for a prefix command.
func (c *PrefixCommand) Aliases(names ...string) *PrefixCommand {
	if c == nil || c.router == nil {
		return c
	}
	c.router.mu.Lock()
	defer c.router.mu.Unlock()
	for _, name := range names {
		name = normalizePrefixName(name)
		if name != "" {
			c.router.prefixCommands[name] = c
		}
	}
	return c
}

// Router manages slash and prefix command registration and dispatching.
type Router struct {
	mu             sync.RWMutex
	commands       map[string]*Command
	prefixCommands map[string]*PrefixCommand
	middleware     []Middleware
	errorHandler   func(error)
	buttons        map[string]*InteractionRoute
	buttonPrefixes []*InteractionRoute
	selects        map[string]*InteractionRoute
	selectPrefixes []*InteractionRoute
	modals         map[string]*InteractionRoute
	modalPrefixes  []*InteractionRoute
	autocomplete   map[string]*InteractionRoute
}

// NewRouter creates an empty command router.
func NewRouter() *Router {
	return &Router{
		commands:       make(map[string]*Command),
		prefixCommands: make(map[string]*PrefixCommand),
		buttons:        make(map[string]*InteractionRoute),
		selects:        make(map[string]*InteractionRoute),
		modals:         make(map[string]*InteractionRoute),
		autocomplete:   make(map[string]*InteractionRoute),
	}
}

// Use adds global middleware that applies to all slash commands.
func (r *Router) Use(mw Middleware) {
	if r == nil || mw == nil {
		return
	}
	r.mu.Lock()
	r.middleware = append(r.middleware, mw)
	r.mu.Unlock()
}

// Command registers a slash command. For startup validation, use CommandE or
// Validate before starting the bot.
func (r *Router) Command(name, description string, handler CommandHandler, opts ...interactions.ApplicationCommandOption) *Command {
	if r == nil {
		return nil
	}
	name = strings.ToLower(strings.TrimSpace(name))
	command := &Command{Name: name, Description: description, Type: interactions.ApplicationCommandTypeChatInput, Options: append([]interactions.ApplicationCommandOption(nil), opts...), Handler: handler}
	r.mu.Lock()
	duplicate := r.commands[name] != nil
	r.commands[name] = command
	r.mu.Unlock()
	if duplicate {
		r.reportError(fmt.Errorf("bot: duplicate command %q replaced", name))
	}
	return command
}

// ContextCommand registers a user or message context-menu command.
func (r *Router) ContextCommand(name string, typ interactions.ApplicationCommandType, handler CommandHandler) *Command {
	if r == nil {
		return nil
	}
	if typ != interactions.ApplicationCommandTypeUser && typ != interactions.ApplicationCommandTypeMessage {
		r.reportError(fmt.Errorf("context command %q has unsupported type %d", name, typ))
		return nil
	}
	name = strings.ToLower(strings.TrimSpace(name))
	command := &Command{Name: name, Type: typ, Handler: handler}
	r.mu.Lock()
	duplicate := r.commands[strings.ToLower(name)] != nil
	r.commands[strings.ToLower(name)] = command
	r.mu.Unlock()
	if duplicate {
		r.reportError(fmt.Errorf("bot: duplicate context command %q replaced", name))
	}
	return command
}

// UserCommand registers a user context-menu command.
func (r *Router) UserCommand(name string, handler CommandHandler) *Command {
	return r.ContextCommand(name, interactions.ApplicationCommandTypeUser, handler)
}

// MessageCommand registers a message context-menu command.
func (r *Router) MessageCommand(name string, handler CommandHandler) *Command {
	return r.ContextCommand(name, interactions.ApplicationCommandTypeMessage, handler)
}

func (r *Router) interactionRoute(id string, handler InteractionHandler) *InteractionRoute {
	return &InteractionRoute{ID: id, Handler: handler}
}

// Button registers an exact custom ID button handler.
func (r *Router) Button(customID string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(customID, handler)
	r.mu.Lock()
	r.buttons[customID] = route
	r.mu.Unlock()
	return route
}

// ButtonPrefix registers a button handler for IDs beginning with prefix.
func (r *Router) ButtonPrefix(prefix string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(prefix, handler)
	r.mu.Lock()
	r.buttonPrefixes = append(r.buttonPrefixes, route)
	r.mu.Unlock()
	return route
}

// Select registers an exact custom ID select-menu handler.
func (r *Router) Select(customID string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(customID, handler)
	r.mu.Lock()
	r.selects[customID] = route
	r.mu.Unlock()
	return route
}

// SelectPrefix registers a select-menu handler for IDs beginning with prefix.
func (r *Router) SelectPrefix(prefix string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(prefix, handler)
	r.mu.Lock()
	r.selectPrefixes = append(r.selectPrefixes, route)
	r.mu.Unlock()
	return route
}

// Modal registers an exact custom ID modal handler.
func (r *Router) Modal(customID string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(customID, handler)
	r.mu.Lock()
	r.modals[customID] = route
	r.mu.Unlock()
	return route
}

// ModalPrefix registers a modal handler for IDs beginning with prefix.
// This is useful for dynamic modal IDs like `supreq_stop_modal_<requestID>`
// where the request ID is appended at runtime.
func (r *Router) ModalPrefix(prefix string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(prefix, handler)
	r.mu.Lock()
	r.modalPrefixes = append(r.modalPrefixes, route)
	r.mu.Unlock()
	return route
}

// Autocomplete registers a handler for a slash command's autocomplete events.
func (r *Router) Autocomplete(command string, handler InteractionHandler) *InteractionRoute {
	route := r.interactionRoute(strings.ToLower(strings.TrimSpace(command)), handler)
	r.mu.Lock()
	r.autocomplete[route.ID] = route
	r.mu.Unlock()
	return route
}

// CommandE registers a slash command and returns validation errors immediately.
func (r *Router) CommandE(name, description string, handler CommandHandler, opts ...interactions.ApplicationCommandOption) (*Command, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if err := validateCommand(name, description, handler, opts); err != nil {
		return nil, err
	}
	if r.HasCommand(name) {
		return nil, fmt.Errorf("bot: duplicate command %q", name)
	}
	return r.Command(name, description, handler, opts...), nil
}

// MustCommand registers a slash command or panics with a useful validation
// error. It is convenient for static command setup in main.
func (r *Router) MustCommand(name, description string, handler CommandHandler, opts ...interactions.ApplicationCommandOption) *Command {
	command, err := r.CommandE(name, description, handler, opts...)
	if err != nil {
		panic(err)
	}
	return command
}

// RemoveCommand removes a slash command by name.
func (r *Router) RemoveCommand(name string) bool {
	if r == nil {
		return false
	}
	name = strings.ToLower(strings.TrimSpace(name))
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.commands[name]; !ok {
		return false
	}
	delete(r.commands, name)
	return true
}

// Lookup returns a slash or context-menu command by name.
func (r *Router) Lookup(name string) (*Command, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	command, ok := r.commands[strings.ToLower(strings.TrimSpace(name))]
	r.mu.RUnlock()
	return command, ok
}

// HasCommand reports whether a command is registered.
func (r *Router) HasCommand(name string) bool {
	_, ok := r.Lookup(name)
	return ok
}

// CommandCount returns the number of registered slash and context-menu commands.
func (r *Router) CommandCount() int {
	if r == nil {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.commands)
}

// RangeCommands iterates over commands in stable name order.
func (r *Router) RangeCommands(fn func(*Command) bool) {
	if fn == nil {
		return
	}
	for _, command := range r.Commands() {
		if !fn(command) {
			return
		}
	}
}

// Prefix registers a text-based prefix command.
func (r *Router) Prefix(name string, handler PrefixHandler) *PrefixCommand {
	if r == nil {
		return nil
	}
	name = normalizePrefixName(name)
	command := &PrefixCommand{Name: name, Handler: handler, router: r}
	r.mu.Lock()
	r.prefixCommands[name] = command
	r.mu.Unlock()
	return command
}

// RemovePrefix removes a prefix command by name or alias.
func (r *Router) RemovePrefix(name string) bool {
	if r == nil {
		return false
	}
	name = normalizePrefixName(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.prefixCommands[name]; !ok {
		return false
	}
	delete(r.prefixCommands, name)
	return true
}

// Commands returns a stable, name-sorted view of registered slash commands.
func (r *Router) Commands() []*Command {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	commands := make([]*Command, 0, len(r.commands))
	for _, command := range r.commands {
		commands = append(commands, command)
	}
	r.mu.RUnlock()
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	return commands
}

// Validate checks all registered slash commands and returns one combined error.
func (r *Router) Validate() error {
	if r == nil {
		return errors.New("bot: router is nil")
	}
	r.mu.RLock()
	commands := make([]*Command, 0, len(r.commands))
	for _, command := range r.commands {
		commands = append(commands, command)
	}
	r.mu.RUnlock()
	sort.Slice(commands, func(i, j int) bool { return commands[i].Name < commands[j].Name })
	var problems []string
	for _, command := range commands {
		command.mu.RLock()
		err := validateRegisteredCommand(command)
		command.mu.RUnlock()
		if err != nil {
			problems = append(problems, err.Error())
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func (r *Router) slashCommands() ([]rest.CreateCommandParams, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	commands := r.Commands()
	result := make([]rest.CreateCommandParams, 0, len(commands))
	for _, command := range commands {
		command.mu.RLock()
		typ := command.Type
		result = append(result, rest.CreateCommandParams{Name: command.Name, Type: &typ, Description: command.Description, Options: append([]interactions.ApplicationCommandOption(nil), command.Options...)})
		command.mu.RUnlock()
	}
	return result, nil
}

// buildSlashCommands is retained internally for simple integrations.
func (r *Router) buildSlashCommands() []rest.CreateCommandParams {
	commands, _ := r.slashCommands()
	return commands
}

// handleInteraction routes an interaction. It returns false when no command
// matches, allowing callers to expose useful metrics.
func (r *Router) handleInteraction(ctx *InteractionContext) bool {
	if r == nil || ctx == nil {
		return false
	}
	r.mu.RLock()
	global := append([]Middleware(nil), r.middleware...)
	var route *InteractionRoute
	var command *Command
	var ok bool
	switch {
	case ctx.Type == interactions.InteractionTypeApplicationCommand:
		command, ok = r.commands[ctx.CommandName()]
	case ctx.IsAutocomplete():
		route = r.autocomplete[ctx.CommandName()]
		ok = route != nil
	case ctx.IsButton():
		route = r.buttons[ctx.CustomID()]
		if route == nil {
			route = matchPrefix(r.buttonPrefixes, ctx.CustomID())
			if route != nil {
				ctx.matchedPrefix = route.ID
			}
		}
		ok = route != nil
	case ctx.IsSelectMenu():
		route = r.selects[ctx.CustomID()]
		if route == nil {
			route = matchPrefix(r.selectPrefixes, ctx.CustomID())
			if route != nil {
				ctx.matchedPrefix = route.ID
			}
		}
		ok = route != nil
	case ctx.IsModalSubmit():
		route = r.modals[ctx.CustomID()]
		if route == nil {
			route = matchPrefix(r.modalPrefixes, ctx.CustomID())
			if route != nil {
				ctx.matchedPrefix = route.ID
			}
		}
		ok = route != nil
	}
	r.mu.RUnlock()
	if !ok {
		return false
	}
	var handler CommandHandler
	var specific []Middleware
	if command != nil {
		command.mu.RLock()
		handler = command.Handler
		specific = append([]Middleware(nil), command.middleware...)
		command.mu.RUnlock()
	} else {
		route.mu.RLock()
		if route.Handler != nil {
			handler = CommandHandler(route.Handler)
		}
		specific = append([]Middleware(nil), route.middleware...)
		route.mu.RUnlock()
	}
	name := ctx.CommandName()
	if command == nil {
		name = route.ID
	}
	if handler == nil {
		r.reportError(fmt.Errorf("command %q has no handler", name))
		return true
	}

	for index := len(specific) - 1; index >= 0; index-- {
		handler = specific[index](handler)
	}
	for index := len(global) - 1; index >= 0; index-- {
		handler = global[index](handler)
	}
	if handler == nil {
		r.reportError(fmt.Errorf("command %q middleware returned a nil handler", name))
		return true
	}
	handler(ctx)
	return true
}

// handlePrefix checks if a message matches a registered prefix command.
func (r *Router) handlePrefix(ctx *MessageContext, prefix string) bool {
	if r == nil || ctx == nil || !strings.HasPrefix(ctx.Content, prefix) {
		return false
	}
	return r.handlePrefixContent(ctx, strings.TrimPrefix(ctx.Content, prefix), prefix)
}

func (r *Router) handlePrefixContent(ctx *MessageContext, content, prefix string) bool {
	if r == nil || ctx == nil {
		return false
	}
	parts, err := parsePrefixArgs(content)
	if err != nil {
		r.reportError(fmt.Errorf("parse prefix command: %w", err))
		return false
	}
	if len(parts) == 0 {
		return false
	}
	name := strings.ToLower(parts[0])
	r.mu.RLock()
	command, ok := r.prefixCommands[name]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	command.mu.RLock()
	handler := command.Handler
	middleware := append([]PrefixMiddleware(nil), command.middleware...)
	minArgs := command.minArgs
	usage := command.usage
	validations := append([]PrefixValidation(nil), command.validations...)
	command.mu.RUnlock()
	if handler == nil {
		r.reportError(fmt.Errorf("prefix command %q has no handler", name))
		return true
	}
	for index := len(middleware) - 1; index >= 0; index-- {
		handler = middleware[index](handler)
	}
	if handler == nil {
		r.reportError(fmt.Errorf("prefix command %q middleware returned a nil handler", name))
		return true
	}
	args := append([]string(nil), parts[1:]...)
	if len(args) < minArgs {
		message := "Usage: " + prefix + command.Name
		if usage != "" {
			message += " " + usage
		}
		_, _ = ctx.Reply(message)
		return true
	}
	for _, validate := range validations {
		if err := validate(ctx, args); err != nil {
			_, _ = ctx.Reply(err.Error())
			return true
		}
	}
	handler(ctx, args)
	return true
}

func (r *Router) setErrorHandler(handler func(error)) {
	r.mu.Lock()
	r.errorHandler = handler
	r.mu.Unlock()
}

func (r *Router) reportError(err error) {
	if err == nil {
		return
	}
	r.mu.RLock()
	handler := r.errorHandler
	r.mu.RUnlock()
	if handler != nil {
		handler(err)
	}
}

func validateCommand(name, description string, handler CommandHandler, options []interactions.ApplicationCommandOption) error {
	return validateCommandType(name, description, interactions.ApplicationCommandTypeChatInput, handler, options)
}

func validateRegisteredCommand(command *Command) error {
	return validateCommandType(command.Name, command.Description, command.Type, command.Handler, command.Options)
}

func validateCommandType(name, description string, typ interactions.ApplicationCommandType, handler CommandHandler, options []interactions.ApplicationCommandOption) error {
	if name == "" || len(name) > 32 {
		return fmt.Errorf("command %q has an invalid name", name)
	}
	for _, char := range name {
		if !(unicode.IsLower(char) || unicode.IsDigit(char) || char == '-' || char == '_') {
			return fmt.Errorf("command %q has an invalid name", name)
		}
	}
	if typ == interactions.ApplicationCommandTypeChatInput && (strings.TrimSpace(description) == "" || len(description) > 100) {
		return fmt.Errorf("command %q has an invalid description", name)
	}
	if typ != interactions.ApplicationCommandTypeChatInput && typ != interactions.ApplicationCommandTypeUser && typ != interactions.ApplicationCommandTypeMessage {
		return fmt.Errorf("command %q has an invalid type", name)
	}
	if typ != interactions.ApplicationCommandTypeChatInput && len(options) > 0 {
		return fmt.Errorf("context command %q cannot have options", name)
	}
	if handler == nil {
		return fmt.Errorf("command %q has no handler", name)
	}
	if err := validateCommandOptions(name, options, false); err != nil {
		return err
	}
	return nil
}

// validateCommandOptions validates one level of command options, recursing
// into subcommand and subcommand group options. inSubcommand is true once
// we are inside a subcommand (where further grouping is forbidden).
func validateCommandOptions(cmdName string, options []interactions.ApplicationCommandOption, inSubcommand bool) error {
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		optionName := strings.ToLower(strings.TrimSpace(option.Name))
		if optionName == "" || len(optionName) > 32 {
			return fmt.Errorf("command %q has an option with an invalid name", cmdName)
		}
		if strings.TrimSpace(option.Description) == "" || len(option.Description) > 100 {
			return fmt.Errorf("command %q option %q has an invalid description", cmdName, option.Name)
		}
		if _, exists := seen[optionName]; exists {
			return fmt.Errorf("command %q has duplicate option %q", cmdName, option.Name)
		}
		seen[optionName] = struct{}{}

		grouping := option.Type == interactions.ApplicationCommandOptionTypeSubCommand ||
			option.Type == interactions.ApplicationCommandOptionTypeSubCommandGroup
		if grouping && option.Required {
			return fmt.Errorf("command %q subcommand %q cannot be required", cmdName, option.Name)
		}
		if grouping {
			if inSubcommand {
				return fmt.Errorf("command %q cannot nest subcommand group %q inside subcommand", cmdName, option.Name)
			}
			// A group must contain at least one subcommand; a plain
			// subcommand may have zero options (e.g. "/giveaway list").
			if option.Type == interactions.ApplicationCommandOptionTypeSubCommandGroup {
				if len(option.Options) == 0 {
					return fmt.Errorf("command %q subcommand group %q has no subcommands", cmdName, option.Name)
				}
				for _, child := range option.Options {
					if child.Type != interactions.ApplicationCommandOptionTypeSubCommand {
						return fmt.Errorf("command %q group %q contains a non-subcommand option %q", cmdName, option.Name, child.Name)
					}
				}
			}
			if len(option.Options) > 0 {
				if err := validateCommandOptions(cmdName, option.Options, true); err != nil {
					return err
				}
			}
			continue
		}
		if len(option.Options) > 0 {
			return fmt.Errorf("command %q option %q takes nested options but is not a subcommand", cmdName, option.Name)
		}
	}
	return nil
}

func normalizePrefixName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func parsePrefixArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false
	started := false
	for _, char := range input {
		if escaped {
			current.WriteRune(char)
			escaped = false
			started = true
			continue
		}
		if char == '\\' {
			escaped = true
			started = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			started = true
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			started = true
			continue
		}
		if unicode.IsSpace(char) {
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
			continue
		}
		current.WriteRune(char)
		started = true
	}
	if escaped {
		current.WriteByte('\\')
	}
	if quote != 0 {
		return nil, errors.New("unterminated quote")
	}
	if started {
		args = append(args, current.String())
	}
	return args, nil
}

// matchPrefix finds the longest matching prefix route for the given custom ID.
// This ensures that overlapping prefixes (e.g. "supreq_cost_" and
// "supreq_cost_done_") resolve to the most specific handler.
func matchPrefix(routes []*InteractionRoute, customID string) *InteractionRoute {
	var best *InteractionRoute
	bestLen := -1
	for _, candidate := range routes {
		if strings.HasPrefix(customID, candidate.ID) && len(candidate.ID) > bestLen {
			best = candidate
			bestLen = len(candidate.ID)
		}
	}
	return best
}
