// recorderbot joins the voice channel of a user who runs !record and writes
// received, decrypted Opus frames as per-user WAV files. It records only while
// explicitly enabled with !record and announces the recording in the text channel.
package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/discord-go/discord.go/client"
	"github.com/discord-go/discord.go/events"
	"github.com/discord-go/discord.go/gateway"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/voice"
	"github.com/gorilla/websocket"
	"layeh.com/gopus"
)

type wsConnection struct{ conn *websocket.Conn }

func (w *wsConnection) Read() ([]byte, error)             { _, b, e := w.conn.ReadMessage(); return b, e }
func (w *wsConnection) ReadMessage() (int, []byte, error) { return w.conn.ReadMessage() }
func (w *wsConnection) Write(b []byte) error              { return w.conn.WriteMessage(websocket.TextMessage, b) }
func (w *wsConnection) WriteBinary(b []byte) error {
	return w.conn.WriteMessage(websocket.BinaryMessage, b)
}
func (w *wsConnection) Close() error { return w.conn.Close() }

type voiceState struct {
	UserID    snowflake.ID  `json:"user_id,string"`
	ChannelID *snowflake.ID `json:"channel_id,string"`
	SessionID string        `json:"session_id"`
}

type voiceServerUpdate struct {
	Token    string       `json:"token"`
	GuildID  snowflake.ID `json:"guild_id,string"`
	Endpoint string       `json:"endpoint"`
}

type recorder struct {
	mu     sync.Mutex
	active bool
	dir    string
	files  map[string]*wavFile
	voices map[snowflake.ID]*voice.Client
}

type wavFile struct {
	file      *os.File
	decoder   *gopus.Decoder
	dataBytes uint32
}

func newRecorder(dir string) *recorder {
	return &recorder{dir: dir, files: make(map[string]*wavFile), voices: make(map[snowflake.ID]*voice.Client)}
}

func (r *recorder) start() error {
	if err := os.MkdirAll(r.dir, 0o700); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.active {
		return fmt.Errorf("recording is already active")
	}
	r.active = true
	return nil
}

func (r *recorder) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.active = false
	for id, wf := range r.files {
		_ = finishWAV(wf)
		delete(r.files, id)
	}
	for guildID, vc := range r.voices {
		_ = vc.Disconnect()
		delete(r.voices, guildID)
	}
}

func (r *recorder) setVoice(guildID snowflake.ID, vc *voice.Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		_ = vc.Disconnect()
		return
	}
	r.voices[guildID] = vc
}

func (r *recorder) write(userID string, opus []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.active {
		return nil
	}
	wf := r.files[userID]
	if wf == nil {
		f, err := os.OpenFile(filepath.Join(r.dir, userID+".wav"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		dec, err := gopus.NewDecoder(48000, 2)
		if err != nil {
			_ = f.Close()
			return err
		}
		wf = &wavFile{file: f, decoder: dec}
		if err := writeWAVHeader(f, 0); err != nil {
			_ = f.Close()
			return err
		}
		if _, err := f.Seek(44, 0); err != nil {
			_ = f.Close()
			return err
		}
		r.files[userID] = wf
	}
	pcm, err := wf.decoder.Decode(opus, 5760, false)
	if err != nil {
		return err
	}
	buf := make([]byte, len(pcm)*2)
	for i, sample := range pcm {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(sample))
	}
	n, err := wf.file.Write(buf)
	wf.dataBytes += uint32(n)
	return err
}

func writeWAVHeader(f *os.File, dataSize uint32) error {
	h := make([]byte, 44)
	copy(h[0:4], "RIFF")
	binary.LittleEndian.PutUint32(h[4:8], 36+dataSize)
	copy(h[8:12], "WAVEfmt ")
	binary.LittleEndian.PutUint32(h[16:20], 16)
	binary.LittleEndian.PutUint16(h[20:22], 1)
	binary.LittleEndian.PutUint16(h[22:24], 2)
	binary.LittleEndian.PutUint32(h[24:28], 48000)
	binary.LittleEndian.PutUint32(h[28:32], 192000)
	binary.LittleEndian.PutUint16(h[32:34], 4)
	binary.LittleEndian.PutUint16(h[34:36], 16)
	copy(h[36:40], "data")
	binary.LittleEndian.PutUint32(h[40:44], dataSize)
	_, err := f.WriteAt(h, 0)
	return err
}

func finishWAV(wf *wavFile) error {
	if err := writeWAVHeader(wf.file, wf.dataBytes); err != nil {
		_ = wf.file.Close()
		return err
	}
	return wf.file.Close()
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}
	outDir := os.Getenv("RECORDINGS_DIR")
	if outDir == "" {
		outDir = "recordings"
	}
	testGuildID, testChannelID, testTextID := testVoiceTarget()

	bot := client.New(token, client.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent|intents.GuildVoiceStates))
	rec := newRecorder(outDir)
	var mu sync.RWMutex
	channels := make(map[snowflake.ID]snowflake.ID) // user -> channel
	sessions := make(map[snowflake.ID]string)       // user -> voice session
	var ownID snowflake.ID

	bot.Gateway.Dispatcher.AddHandler(func(raw []byte) {
		var p gateway.GatewayPayload
		if json.Unmarshal(raw, &p) != nil || p.Type == nil {
			return
		}
		switch *p.Type {
		case "READY":
			var d struct {
				User struct {
					ID snowflake.ID `json:"id,string"`
				} `json:"user"`
			}
			if json.Unmarshal(p.Data, &d) == nil {
				mu.Lock()
				ownID = d.User.ID
				mu.Unlock()
				if testGuildID != 0 {
					if err := rec.start(); err != nil {
						log.Printf("voice test recorder start: %v", err)
						return
					}
					send(bot, testTextID, "Voice test recording started; synthetic music-bot tone only.")
					if err := bot.Gateway.JoinVoiceChannel(testGuildID, testChannelID, false, false); err != nil {
						log.Printf("voice test recorder join: %v", err)
						rec.stop()
					}
				}
			}
		case "GUILD_CREATE":
			var d struct {
				VoiceStates []voiceState `json:"voice_states"`
			}
			if json.Unmarshal(p.Data, &d) == nil {
				mu.Lock()
				for _, vs := range d.VoiceStates {
					if vs.ChannelID != nil {
						channels[vs.UserID] = *vs.ChannelID
					}
					if vs.SessionID != "" {
						sessions[vs.UserID] = vs.SessionID
					}
				}
				mu.Unlock()
			}
		case "VOICE_STATE_UPDATE":
			var vs voiceState
			if json.Unmarshal(p.Data, &vs) == nil {
				mu.Lock()
				if vs.ChannelID == nil {
					delete(channels, vs.UserID)
				} else {
					channels[vs.UserID] = *vs.ChannelID
				}
				if vs.SessionID != "" {
					sessions[vs.UserID] = vs.SessionID
				}
				mu.Unlock()
			}
		case "VOICE_SERVER_UPDATE":
			var vsu voiceServerUpdate
			if json.Unmarshal(p.Data, &vsu) == nil {
				mu.RLock()
				id := ownID
				mu.RUnlock()
				go connectRecorder(rec, vsu, id, &mu, sessions, channels)
			}
		case "MESSAGE_CREATE":
			var e events.MessageCreate
			if json.Unmarshal(p.Data, &e) != nil || e.Author == nil || e.Author.Bot {
				return
			}
			cmd := strings.TrimSpace(e.Content)
			switch cmd {
			case "!record":
				mu.RLock()
				channelID, ok := channels[e.Author.ID]
				mu.RUnlock()
				if !ok {
					send(bot, e.ChannelID, "Join a voice channel first, then run `!record`.")
					return
				}
				if err := rec.start(); err != nil {
					send(bot, e.ChannelID, err.Error())
					return
				}
				if err := bot.Gateway.JoinVoiceChannel(e.GuildID, channelID, false, false); err != nil {
					rec.stop()
					send(bot, e.ChannelID, "Could not join voice: "+err.Error())
					return
				}
				send(bot, e.ChannelID, "Recording is on. Run `!stoprecord` to stop and save WAV files.")
			case "!stoprecord":
				rec.stop()
				_ = bot.Gateway.JoinVoiceChannel(e.GuildID, 0, false, false)
				send(bot, e.ChannelID, "Recording stopped. WAV files are saved in the recordings directory.")
			}
		}
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	log.Printf("Starting recorder bot; output directory: %s", outDir)
	if err := bot.Gateway.Start(ctx); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
	<-ctx.Done()
	rec.stop()
}

func testVoiceTarget() (snowflake.ID, snowflake.ID, snowflake.ID) {
	guildRaw := os.Getenv("VOICE_TEST_GUILD_ID")
	channelRaw := os.Getenv("VOICE_TEST_CHANNEL_ID")
	textRaw := os.Getenv("VOICE_TEST_TEXT_CHANNEL_ID")
	if guildRaw == "" && channelRaw == "" && textRaw == "" {
		return 0, 0, 0
	}
	guildID, err := snowflake.Parse(guildRaw)
	if err != nil {
		log.Fatalf("invalid VOICE_TEST_GUILD_ID: %v", err)
	}
	channelID, err := snowflake.Parse(channelRaw)
	if err != nil {
		log.Fatalf("invalid VOICE_TEST_CHANNEL_ID: %v", err)
	}
	textID, err := snowflake.Parse(textRaw)
	if err != nil {
		log.Fatalf("invalid VOICE_TEST_TEXT_CHANNEL_ID: %v", err)
	}
	return guildID, channelID, textID
}

func send(bot *client.Client, channelID snowflake.ID, content string) {
	if _, err := bot.Rest.CreateMessage(context.Background(), channelID, content); err != nil {
		log.Printf("send status message: %v", err)
	}
}

func connectRecorder(rec *recorder, vsu voiceServerUpdate, botID snowflake.ID, mu *sync.RWMutex, sessions map[snowflake.ID]string, channels map[snowflake.ID]snowflake.ID) {
	var sessionID string
	var channelID snowflake.ID
	for range 100 {
		mu.RLock()
		sessionID = sessions[botID]
		channelID = channels[botID]
		mu.RUnlock()
		if sessionID != "" && channelID != 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if sessionID == "" || channelID == 0 {
		log.Printf("voice connect: bot voice state is not available")
		return
	}
	endpoint := vsu.Endpoint
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "wss://"), "ws://")
	endpoint = "wss://" + strings.TrimSuffix(endpoint, "/")
	if !strings.Contains(endpoint, "?") {
		endpoint += "/?v=4"
	}
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Printf("voice websocket dial: %v", err)
		return
	}
	vc := voice.NewClient(&wsConnection{conn: conn}, vsu.GuildID, channelID, botID)
	vc.SessionID, vc.Token, vc.Endpoint = sessionID, vsu.Token, endpoint
	vc.OnAudioPacket = func(userID string, _ uint16, _ uint32, opus []byte) {
		if err := rec.write(userID, opus); err != nil {
			log.Printf("decode/write audio for user %s: %v", userID, err)
		}
	}
	if err := vc.Connect(context.Background()); err != nil {
		log.Printf("voice handshake: %v", err)
		_ = conn.Close()
		return
	}
	rec.setVoice(vsu.GuildID, vc)
	log.Printf("Recorder voice connection established")
}
