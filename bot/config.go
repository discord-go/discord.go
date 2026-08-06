package bot

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/discord-go/discord.go/intents"
)

// Config contains application-owned bot startup settings. It intentionally
// contains no secrets beyond the token and can be loaded from JSON or env.
type Config struct {
	Token           string             `json:"token,omitempty"`
	Prefix          string             `json:"prefix,omitempty"`
	BotName         string             `json:"bot_name,omitempty"`
	MentionTriggers bool               `json:"mention_triggers,omitempty"`
	Intents         intents.Intent     `json:"intents,omitempty"`
	Shards          int                `json:"shards,omitempty"`
	AutomaticShards bool               `json:"automatic_shards,omitempty"`
	Compression     bool               `json:"compression,omitempty"`
	Presence        *PresenceUpdate    `json:"presence,omitempty"`
	CommandSync     *CommandSyncConfig `json:"command_sync,omitempty"`
}

// LoadConfig loads bot settings from a JSON file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

// ConfigFromEnv reads the conventional TOKEN/DISCORD_TOKEN and optional bot
// trigger settings from environment variables.
func ConfigFromEnv() Config {
	token := os.Getenv("TOKEN")
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
	}
	config := Config{Token: token, Prefix: os.Getenv("BOT_PREFIX"), BotName: os.Getenv("BOT_NAME")}
	config.MentionTriggers, _ = strconv.ParseBool(os.Getenv("BOT_MENTION_TRIGGERS"))
	if value, err := strconv.ParseInt(os.Getenv("BOT_INTENTS"), 10, 64); err == nil {
		config.Intents = intents.Intent(value)
	}
	if value, err := strconv.Atoi(os.Getenv("BOT_SHARDS")); err == nil {
		config.Shards = value
	}
	config.AutomaticShards, _ = strconv.ParseBool(os.Getenv("BOT_AUTOMATIC_SHARDS"))
	config.Compression, _ = strconv.ParseBool(os.Getenv("BOT_GATEWAY_COMPRESSION"))
	return config
}

// NewFromConfig creates a Bot from application configuration.
func NewFromConfig(config Config, opts ...Option) *Bot {
	if config.Prefix != "" {
		opts = append(opts, WithPrefix(config.Prefix))
	}
	if config.BotName != "" {
		opts = append(opts, WithBotName(config.BotName))
	}
	if config.MentionTriggers {
		opts = append(opts, WithMentionTriggers(true))
	}
	if config.Intents != 0 {
		opts = append(opts, WithIntents(config.Intents))
	}
	if config.AutomaticShards || config.Shards > 0 {
		opts = append(opts, WithShards(config.Shards))
	}
	if config.Compression {
		opts = append(opts, WithGatewayCompression(true))
	}
	if config.Presence != nil {
		opts = append(opts, WithPresence(*config.Presence))
	}
	if config.CommandSync != nil {
		opts = append(opts, WithCommandSync(*config.CommandSync))
	}
	return New(config.Token, opts...)
}
