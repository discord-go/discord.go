// discord.go code
package main

import (
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		token = os.Getenv("DISCORD_TOKEN")
	}
	if token == "" {
		log.Fatal("TOKEN or DISCORD_TOKEN is required")
	}

	config := loadTemplateConfig()
	router := bot.NewRouter()
	router.Use(bot.GuildOnly())
	registerSlashCommands(router, config)
	registerMessageCommands(router, config)

	client := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMembers|intents.GuildMessages|intents.GuildMessageReactions|intents.DirectMessages|intents.MessageContent|intents.GuildVoiceStates),
		bot.WithPrefix(config.Prefix),
		bot.WithBotName(config.BotName),
		bot.WithMentionTriggers(true),
		bot.WithRouter(router),
		bot.WithShards(0),
		bot.WithPresence(bot.PresenceUpdate{Status: "online", Activities: []bot.Activity{{Name: "Components V2", Type: 4}}}),
	)
	registerTemplateEvents(client)
	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}
