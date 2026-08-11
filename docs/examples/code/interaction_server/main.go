// Package main demonstrates a Discord interaction HTTP server using
// interactions.Server, which verifies the Ed25519 signature and timestamp
// freshness of every incoming request automatically.
//
// This example illustrates:
//  1. Creating an interactions.Server with your application's public key
//  2. Handling PING (type 1) automatically — the server responds with PONG
//  3. Handling an application command (type 2) interaction
//  4. Running the server on localhost for development
//
// Set DISCORD_PUBLIC_KEY to your application's public key from the Discord
// Developer Portal (General Information > Public Key).
//
// Run: go run ./docs/examples/code/interaction_server
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/discord-go/discord.go/interactions"
)

func main() {
	publicKey := os.Getenv("DISCORD_PUBLIC_KEY")
	if publicKey == "" {
		log.Fatal("DISCORD_PUBLIC_KEY is required (from the Discord Developer Portal)")
	}

	// NewServer creates an http.Handler that:
	//   - reads the request body
	//   - verifies the Ed25519 signature AND timestamp freshness (5-min window)
	//     using interactions.VerifyRequest (NOT VerifySignature, which skips
	//     the timestamp check and allows replay attacks)
	//   - auto-responds to PING (type 1) with PONG
	//   - dispatches all other interactions to your handler
	srv := interactions.NewServer(publicKey, func(i *interactions.Interaction) *interactions.InteractionResponse {
		// Only application command invocations reach here; pings are handled
		// automatically by the server.
		if i.Type != interactions.InteractionTypeApplicationCommand {
			return nil // let the server auto-defer
		}

		// Respond with a simple message. In a real bot you would unmarshal
		// i.Data (json.RawMessage) into interactions.ApplicationCommandData
		// and dispatch on the command name.
		return &interactions.InteractionResponse{
			Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
			Data: &interactions.InteractionCallbackData{
				Content: fmt.Sprintf("Hello, %s!", i.Member.User.Username),
			},
		}
	})

	// Register the server on your HTTP mux. Discord sends POST requests to
	// the URL you configured in the Developer Portal under
	// General Information > Interactions Endpoint URL.
	mux := http.NewServeMux()
	mux.Handle("/interactions", srv)

	addr := ":8080"
	log.Printf("Interaction server listening on %s/interactions", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
