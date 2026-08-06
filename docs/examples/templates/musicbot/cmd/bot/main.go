// discord.go code
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"

	"musicbot/internal/commands"
	"musicbot/internal/player"

	"github.com/gorilla/websocket"
)

// voiceWSConnection wraps gorilla/websocket for the voice.Connection interface.
type voiceWSConnection struct {
	conn *websocket.Conn
}

func (w *voiceWSConnection) Read() ([]byte, error) {
	_, msg, err := w.conn.ReadMessage()
	return msg, err
}

func (w *voiceWSConnection) ReadMessage() (int, []byte, error) {
	return w.conn.ReadMessage()
}

func (w *voiceWSConnection) Write(data []byte) error {
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func (w *voiceWSConnection) WriteBinary(data []byte) error {
	return w.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (w *voiceWSConnection) Close() error {
	return w.conn.Close()
}

type VoiceState struct {
	UserID    snowflake.ID  `json:"user_id,string"`
	ChannelID *snowflake.ID `json:"channel_id,string"`
	SessionID string        `json:"session_id"`
}

type VoiceServerUpdate struct {
	Token    string       `json:"token"`
	GuildID  snowflake.ID `json:"guild_id,string"`
	Endpoint string       `json:"endpoint"`
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		token = "dummy-token"
	}

	bot := client.New(token, client.WithIntents(
		intents.Guilds|
			intents.GuildMessages|
			intents.MessageContent|
			intents.GuildVoiceStates,
	))

	manager := player.NewQueueManager()

	var voiceStatesMu sync.RWMutex
	voiceStates := make(map[snowflake.ID]snowflake.ID) // userID -> channelID

	// Track session IDs for voice (userID -> sessionID)
	var voiceSessionsMu sync.RWMutex
	voiceSessions := make(map[snowflake.ID]string)

	// Track bot's own user ID
	var botUserID snowflake.ID

	bot.Gateway.Dispatcher.AddHandler(func(data []byte) {
		var payload gateway.GatewayPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return
		}

		// Log all dispatches
		if payload.Type != nil {
			log.Printf("DEBUG: Received event %s", *payload.Type)
		} else {
			log.Printf("DEBUG: Received op=%d", payload.Op)
		}

		if payload.Type == nil {
			return
		}

		// Capture bot user ID from READY
		if *payload.Type == "READY" {
			var readyData struct {
				User struct {
					ID snowflake.ID `json:"id,string"`
				} `json:"user"`
			}
			if err := json.Unmarshal(payload.Data, &readyData); err == nil {
				botUserID = readyData.User.ID
				log.Printf("DEBUG: Bot user ID = %d", botUserID)
			}
		}

		if *payload.Type == "GUILD_CREATE" {
			var guildData struct {
				VoiceStates []VoiceState `json:"voice_states"`
			}
			if err := json.Unmarshal(payload.Data, &guildData); err == nil {
				voiceStatesMu.Lock()
				voiceSessionsMu.Lock()
				for _, vs := range guildData.VoiceStates {
					if vs.ChannelID != nil {
						voiceStates[vs.UserID] = *vs.ChannelID
					}
					if vs.SessionID != "" {
						voiceSessions[vs.UserID] = vs.SessionID
					}
				}
				voiceSessionsMu.Unlock()
				voiceStatesMu.Unlock()
			}
		}

		if *payload.Type == "VOICE_STATE_UPDATE" {
			log.Printf("DEBUG: VOICE_STATE_UPDATE: %s", string(payload.Data))
			var vs VoiceState
			if err := json.Unmarshal(payload.Data, &vs); err == nil {
				voiceStatesMu.Lock()
				if vs.ChannelID == nil {
					delete(voiceStates, vs.UserID)
				} else {
					voiceStates[vs.UserID] = *vs.ChannelID
				}
				voiceStatesMu.Unlock()

				voiceSessionsMu.Lock()
				if vs.SessionID != "" {
					voiceSessions[vs.UserID] = vs.SessionID
				}
				voiceSessionsMu.Unlock()
			}
			return
		}

		if *payload.Type == "VOICE_SERVER_UPDATE" {
			log.Printf("DEBUG: VOICE_SERVER_UPDATE: %s", string(payload.Data))
			var vsu VoiceServerUpdate
			if err := json.Unmarshal(payload.Data, &vsu); err == nil {
				go handleVoiceServerUpdate(bot, manager, vsu, botUserID, voiceSessions, &voiceSessionsMu)
			}
			return
		}

		if *payload.Type == "MESSAGE_CREATE" {
			var msg events.MessageCreate
			if err := json.Unmarshal(payload.Data, &msg); err != nil {
				return
			}
			e := &msg
			if e.Author == nil || e.Author.Bot {
				return
			}
			content := strings.TrimSpace(e.Content)
			ctx := context.Background()

			log.Printf("DEBUG: Received command from %s: %q", e.Author.Username, content)

			if strings.HasPrefix(content, "!play ") {
				voiceStatesMu.RLock()
				vcID, ok := voiceStates[e.Author.ID]
				voiceStatesMu.RUnlock()
				commands.HandlePlay(ctx, bot, e, manager, vcID, ok)
			} else if content == "!join" {
				voiceStatesMu.RLock()
				vcID, ok := voiceStates[e.Author.ID]
				voiceStatesMu.RUnlock()
				commands.HandleJoin(ctx, bot, e, vcID, ok)
			} else if content == "!leave" {
				commands.HandleLeave(ctx, bot, e, manager)
			} else if content == "!stop" {
				commands.HandleStop(ctx, bot, e, manager)
			} else if content == "!queue" {
				commands.HandleQueue(ctx, bot, e, manager)
			} else if content == "!info" {
				commands.HandleInfo(ctx, bot, e)
			}
		}
	})

	log.Println("Starting bot...")
	if err := bot.Gateway.Start(context.Background()); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}

	log.Println("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func handleVoiceServerUpdate(
	bot *client.Client,
	manager *player.QueueManager,
	vsu VoiceServerUpdate,
	botUserID snowflake.ID,
	voiceSessions map[snowflake.ID]string,
	voiceSessionsMu *sync.RWMutex,
) {
	p := manager.GetPlayer(vsu.GuildID)

	// Get the bot's session ID for this voice connection
	voiceSessionsMu.RLock()
	sessionID := voiceSessions[botUserID]
	voiceSessionsMu.RUnlock()

	if sessionID == "" {
		log.Printf("ERROR: No session ID found for bot user %d", botUserID)
		return
	}

	endpoint := vsu.Endpoint
	if !strings.HasPrefix(endpoint, "wss://") {
		endpoint = "wss://" + endpoint + "/?v=4"
	}

	log.Printf("DEBUG: Connecting to voice endpoint: %s (session=%s, token=%s)", endpoint, sessionID, vsu.Token)

	// Dial the voice WebSocket
	wsConn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Printf("ERROR: Failed to dial voice WebSocket: %v", err)
		return
	}

	voiceConn := &voiceWSConnection{conn: wsConn}

	// Create voice client
	vc := voice.NewClient(voiceConn, vsu.GuildID, p.ChannelID, botUserID)
	vc.SessionID = sessionID
	vc.Token = vsu.Token
	vc.Endpoint = endpoint

	// Log incoming audio packets
	vc.OnAudioPacket = func(userID string, sequence uint16, timestamp uint32, audio []byte) {
		// To avoid spamming the console, only log every 100th packet per user
		if sequence%100 == 0 {
			log.Printf("AUDIO: Received decrypted audio packet from User %s (seq=%d, len=%d bytes)", userID, sequence, len(audio))
		}
	}

	// Perform the full voice handshake (Hello → Identify → Ready → UDP Discovery → SelectProtocol → SessionDescription)
	log.Printf("DEBUG: Starting voice handshake...")
	if err := vc.Connect(context.Background()); err != nil {
		log.Printf("ERROR: Voice handshake failed: %v", err)
		wsConn.Close()
		return
	}

	log.Printf("DEBUG: Voice connected successfully! Starting audio playback...")

	// Start playing from the queue
	p.PlayNext(vc)
}
