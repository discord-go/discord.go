package bot_test

import (
	"fmt"
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
)

// The smallest complete bot: a slash command that replies.
//
// This example has no Output comment, so `go test` compiles it without
// running it; a real token would connect to the gateway.
func ExampleNew() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("DISCORD_TOKEN is required")
		return
	}

	router := bot.NewRouter()
	router.Command("ping", "Check bot status", func(ctx *bot.InteractionContext) {
		if err := ctx.Reply("Pong!"); err != nil {
			log.Printf("reply: %v", err)
		}
	})

	client := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
	)
	if err := client.Run(); err != nil {
		log.Fatal(err)
	}
}

// Reading command options with the Option* accessors.
func ExampleInteractionContext_OptionBool() {
	// Inside a command handler registered with router.Command:
	//
	//   router.Command("lock", "Lock the channel", func(ctx *bot.InteractionContext) {
	//       notify := ctx.OptionBool("notify")       // bool option
	//       reason := ctx.OptionString("reason")     // string option
	//       target := ctx.OptionChannel("channel")   // channel option as snowflake.ID
	//       _ = notify
	//       _ = reason
	//       _ = target
	//   })
	//
	// OptionBool, OptionString, OptionInt, OptionFloat, and OptionSnowflake
	// resolve options nested inside subcommands. The former Get* names
	// (GetBool, GetString, Get*Option) remain as deprecated aliases.
	fmt.Println("accessors resolve nested subcommand options")
	// Output: accessors resolve nested subcommand options
}
