package interactions

// SlashCommandBuilder provides a fluent interface for constructing Application Commands.
type SlashCommandBuilder struct {
	cmd ApplicationCommand
}

// NewSlashCommandBuilder creates a new builder for a chat input (slash) command.
func NewSlashCommandBuilder(name, description string) *SlashCommandBuilder {
	cmdType := ApplicationCommandTypeChatInput
	return &SlashCommandBuilder{
		cmd: ApplicationCommand{
			Type:        &cmdType,
			Name:        name,
			Description: description,
		},
	}
}

// SetName sets the command name.
func (b *SlashCommandBuilder) SetName(name string) *SlashCommandBuilder {
	b.cmd.Name = name
	return b
}

// SetDescription sets the command description.
func (b *SlashCommandBuilder) SetDescription(description string) *SlashCommandBuilder {
	b.cmd.Description = description
	return b
}

// AddStringOption adds a string option to the slash command.
func (b *SlashCommandBuilder) AddStringOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeString,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddStringOptionWithChoices adds a string option with predefined choices.
func (b *SlashCommandBuilder) AddStringOptionWithChoices(name, description string, required bool, choices ...ApplicationCommandOptionChoice) *SlashCommandBuilder {
	return b.AddOption(ApplicationCommandOption{Type: ApplicationCommandOptionTypeString, Name: name, Description: description, Required: required, Choices: choices})
}

// AddIntegerOption adds an integer option to the slash command.
func (b *SlashCommandBuilder) AddIntegerOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeInteger,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddBooleanOption adds a boolean option to the slash command.
func (b *SlashCommandBuilder) AddBooleanOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeBoolean,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddUserOption adds a user option to the slash command.
func (b *SlashCommandBuilder) AddUserOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeUser,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddChannelOption adds a channel option to the slash command.
func (b *SlashCommandBuilder) AddChannelOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeChannel,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddRoleOption adds a role option to the slash command.
func (b *SlashCommandBuilder) AddRoleOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeRole,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddMentionableOption adds a mentionable option to the slash command.
func (b *SlashCommandBuilder) AddMentionableOption(name, description string, required bool) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeMentionable,
		Name:        name,
		Description: description,
		Required:    required,
	})
	return b
}

// AddOption adds a raw ApplicationCommandOption to the slash command.
func (b *SlashCommandBuilder) AddOption(opt ApplicationCommandOption) *SlashCommandBuilder {
	b.cmd.Options = append(b.cmd.Options, opt)
	return b
}

// AddSubcommand adds a subcommand option with its own nested options:
//
//	NewSlashCommandBuilder("giveaway", "...").AddSubcommand(
//		"create", "Start a new giveaway",
//		ApplicationCommandOption{Type: Type, Name: "prize", ...},
//	)
func (b *SlashCommandBuilder) AddSubcommand(name, description string, options ...ApplicationCommandOption) *SlashCommandBuilder {
	return b.AddOption(ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeSubCommand,
		Name:        name,
		Description: description,
		Options:     options,
	})
}

// AddSubcommandGroup adds a subcommand group containing subcommands:
//
//	NewSlashCommandBuilder("mod", "...").AddSubcommandGroup(
//		"cases", "Manage moderation cases",
//		ApplicationCommandOption{Type: ApplicationCommandOptionTypeSubCommand, Name: "view", ...},
//	)
func (b *SlashCommandBuilder) AddSubcommandGroup(name, description string, subcommands ...ApplicationCommandOption) *SlashCommandBuilder {
	return b.AddOption(ApplicationCommandOption{
		Type:        ApplicationCommandOptionTypeSubCommandGroup,
		Name:        name,
		Description: description,
		Options:     subcommands,
	})
}

// SetIntegrationTypes sets the integration types (e.g. User Apps V2) for the slash command.
// 0 = Guild Install, 1 = User Install
func (b *SlashCommandBuilder) SetIntegrationTypes(types ...int) *SlashCommandBuilder {
	b.cmd.IntegrationTypes = types
	return b
}

// SetContexts sets the interaction contexts (e.g. DMs, Guilds, Private Channels) for the command.
// 0 = Guild, 1 = Bot DM, 2 = Private Channel
func (b *SlashCommandBuilder) SetContexts(contexts ...int) *SlashCommandBuilder {
	b.cmd.Contexts = contexts
	return b
}

// Build returns the constructed ApplicationCommand.
func (b *SlashCommandBuilder) Build() ApplicationCommand {
	return b.cmd
}
