// discord.go code
package main

import (
	"fmt"
	"log"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/messages"
)

func registerMessageCommands(router *bot.Router, config templateConfig) {
	router.Prefix("ping", func(ctx *bot.MessageContext, _ []string) {
		if ctx.GuildID == 0 {
			_, _ = ctx.Reply("This command can only be used in a server.")
			return
		}
		content := fmt.Sprintf("# Pong!\n**WebSocket Ping:** %dms\n**API Response Time:** measured by discord.go", ctx.Bot.GatewayLatency().Milliseconds())
		response := templateStatusContainer(config.accentColor(), content)
		if _, err := ctx.ReplyComplex(messages.MessageSend{Flags: messages.FlagIsComponentsV2, Components: []components.Component{response}}); err != nil {
			log.Printf("message ping: %v", err)
			_, _ = ctx.Reply("Pong!")
		}
	}).Aliases("p")
}
