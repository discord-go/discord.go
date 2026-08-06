# Canvas Alternatives

## Overview

`discord.go` does not depend on a JavaScript canvas runtime. For generated
visuals, use an embed with a stable image URL, Components V2 text and media
components, or Go's standard `image` and `image/png` packages. The example
below renders a small PNG in memory and uploads it as an interaction
attachment.

## Tutorial: Render An Image In Go

1. Draw into an `image.RGBA` using the standard library.
2. Encode it to PNG in a `bytes.Buffer`.
3. Validate the output size before uploading.
4. Reference it with `attachment://filename` in an embed.
5. Upload a matching `rest.File` with `ReplyComplexWithFiles`.

For charts, use a Go chart library outside the bot package, render SVG or PNG,
and keep the renderer bounded. For simple status cards, an embed or Components
V2 layout is often cheaper than generating pixels.

## Complete Runnable Example

Copy to `examples/canvas-alternatives/main.go`, set `DISCORD_TOKEN`, and run it.
Invoke `/status-card`.

```go
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/rest"
)

func renderStatusCard() ([]byte, error) {
	card := image.NewRGBA(image.Rect(0, 0, 600, 180))
	draw.Draw(card, card.Bounds(), &image.Uniform{C: color.RGBA{R: 24, G: 26, B: 32, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(card, image.Rect(32, 32, 568, 70), &image.Uniform{C: color.RGBA{R: 88, G: 101, B: 242, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(card, image.Rect(32, 96, 480, 128), &image.Uniform{C: color.RGBA{R: 67, G: 181, B: 129, A: 255}}, image.Point{}, draw.Src)
	var output bytes.Buffer
	err := png.Encode(&output, card)
	return output.Bytes(), err
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	router := bot.NewRouter()
	router.Command("status-card", "Render a generated status card", func(ctx *bot.InteractionContext) {
		data, err := renderStatusCard()
		if err != nil {
			_ = ctx.ReplyEphemeral("Could not render the image.")
			return
		}
		if err := rest.ValidateFilesSize([][]byte{data}, 0); err != nil {
			_ = ctx.ReplyEphemeral("The generated image is too large.")
			return
		}
		embed := messages.Embed{
			Title:       "Service status",
			Description: "This image was generated with the Go standard library.",
			Color:       0x5865F2,
			Image:       &messages.EmbedImage{URL: "attachment://status.png"},
		}
		if err := embed.Validate(); err != nil {
			log.Printf("validate status embed: %v", err)
			return
		}
		file := rest.FileFromBytes("status.png", data)
		if err := ctx.ReplyComplexWithFiles(&interactions.InteractionCallbackData{
			Embeds: []messages.Embed{embed},
		}, file); err != nil {
			log.Printf("upload status card: %v", err)
		}
	})

	b := bot.New(token, bot.WithIntents(intents.Guilds), bot.WithRouter(router))
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}
}
```

## Choosing A Technique

- Use `messages.Embed` for structured text with a thumbnail or image.
- Use Components V2 for typed layouts, separators, sections, galleries, and
  files.
- Use an attachment when the bot must generate bytes itself.
- Use a dedicated renderer or worker for expensive charts; do not block the
  interaction handler indefinitely.

## Common Mistakes

- Generating a large image before acknowledging a slash command.
- Letting a user choose an arbitrary local path for `FileFromPath`.
- Referencing `attachment://wrong-name.png` when the uploaded file has another
  name.
- Assuming every client renders every image format identically.
- Failing to enforce a size limit before multipart upload.

## Expected Result

`/status-card` uploads an in-memory PNG and displays it in an embed. No canvas
runtime or external image service is needed.

## Related Pages

- [Embeds](embeds.md)
- [Display Components](../interactions/display-components.md)
- [Common Errors](common-errors.md)
