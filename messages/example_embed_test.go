package messages_test

import (
	"fmt"

	"github.com/discord-go/discord.go/messages"
)

// A simple embed in four chained calls.
func ExampleNewEmbedBuilder() {
	embed := messages.NewEmbedBuilder().
		SetTitle("Welcome").
		SetDescription("Thanks for joining the server.").
		SetColorHex("#5865F2").
		SetTimestampNow().
		Build()

	fmt.Println(embed.Title)
	// Output: Welcome
}

// Embeds with fields, an author, and a footer.
func ExampleEmbedBuilder_AddField() {
	embed := messages.NewEmbedBuilder().
		SetAuthorName("ModBot").
		SetTitle("User warned").
		AddInlineField("User", "@alice").
		AddInlineField("Reason", "spam").
		SetFooterText("Case #12").
		Build()

	fmt.Println(len(embed.Fields), embed.Fields[0].Name)
	// Output: 2 User
}
