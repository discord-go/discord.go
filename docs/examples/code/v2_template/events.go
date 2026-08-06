// discord.go code
package main

import (
	"log"

	"github.com/discord-go/discord.go/bot"
)

func registerTemplateEvents(client *bot.Bot) {
	client.OnReady(func(ctx *bot.ReadyContext) {
		log.Printf("BOT READY: logged in as %s#%s", ctx.User.Username, ctx.User.Discriminator)
	})
	client.OnError(func(err error) {
		log.Printf("BOT ERROR: %v", err)
	})
}
