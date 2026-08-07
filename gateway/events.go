package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/discord-go/discord.go/cache"
	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/internal/compression"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

// GatewayPayload represents a raw payload received from the Discord gateway.
type GatewayPayload struct {
	Op       Opcode          `json:"op"`
	Data     json.RawMessage `json:"d"`
	Sequence *int64          `json:"s,omitempty"`
	Type     *string         `json:"t,omitempty"`
}

// HelloData represents the data in a Hello (op 10) payload.
type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// readLoop continuously reads from the Connection, parses opcodes,
// and routes events to the appropriate handler.
func (c *Client) readLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		data, err := c.Conn.Read()
		if err != nil {
			if code, ok := extractCloseCode(err); ok {
				switch code {
				case CloseCodeInvalidSeq, CloseCodeInvalidShard, CloseCodeShardingRequired, CloseCodeInvalidAPIVersion, CloseCodeInvalidIntents:
					// Reconnect with a full re-identify
					return ErrInvalidSession
				case CloseCodeAuthenticationFailed, CloseCodeDisallowedIntents:
					// Not reconnect at all
					return fmt.Errorf("%w: %d", ErrFatalClose, code)
				default:
					// Resume
					return err
				}
			}
			return err
		}
		if c.Compressed {
			if c.compression == nil {
				c.compression = compression.NewStream()
			}
			data, err = c.compression.Write(data)
			if err != nil {
				return err
			}
			if len(data) == 0 {
				continue
			}
		}

		var payload GatewayPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			// If it's not valid JSON, dispatch the raw data for
			// backward compatibility with the existing dispatcher.
			c.Dispatcher.Dispatch(data)
			continue
		}

		// Route based on opcode.
		switch payload.Op {
		case OpcodeDispatch:
			// Update sequence number if present.
			if payload.Sequence != nil && c.Session != nil {
				c.Session.UpdateSequence(*payload.Sequence)
				if c.Heartbeater != nil {
					c.Heartbeater.UpdateSequence(*payload.Sequence)
				}
			}

			if payload.Type != nil && c.Cache != nil {
				switch *payload.Type {
				case "GUILD_CREATE":
					var guild guilds.Guild
					if err := json.Unmarshal(payload.Data, &guild); err == nil && guild.ID != 0 {
						guild.Unavailable = false
						if gc, ok := c.Cache.(cache.GuildCache); ok {
							gc.SetGuild(guild.ID.String(), &guild)
						}
					}
				case "GUILD_DELETE":
					var data struct {
						ID          string `json:"id"`
						Unavailable bool   `json:"unavailable,omitempty"`
					}
					if err := json.Unmarshal(payload.Data, &data); err == nil && data.ID != "" {
						if gc, ok := c.Cache.(cache.GuildCache); ok {
							if data.Unavailable {
								if cached, found := gc.GetGuild(data.ID); found {
									if guild, ok := cached.(*guilds.Guild); ok {
										guild.Unavailable = true
										gc.SetGuild(data.ID, guild)
									}
								}
							} else {
								gc.DeleteGuild(data.ID)
							}
						}
					}
				case "CHANNEL_DELETE":
					var data struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(payload.Data, &data); err == nil && data.ID != "" {
						if cc, ok := c.Cache.(cache.ChannelCache); ok {
							cc.DeleteChannel(data.ID)
						}
					}
				case "ROLE_DELETE":
					var data struct {
						RoleID  string `json:"role_id"`
						GuildID string `json:"guild_id"`
					}
					if err := json.Unmarshal(payload.Data, &data); err == nil && data.RoleID != "" {
						if rc, ok := c.Cache.(cache.RoleCache); ok {
							rc.DeleteRole(data.RoleID)
						}
					}
				case "MESSAGE_DELETE":
					var data struct {
						ID string `json:"id"`
					}
					if err := json.Unmarshal(payload.Data, &data); err == nil && data.ID != "" {
						if mc, ok := c.Cache.(cache.MessageCache); ok {
							mc.DeleteMessage(data.ID)
						}
					}
				case "PRESENCE_UPDATE":
					var p struct {
						User    users.User `json:"user"`
						GuildID string     `json:"guild_id"`
						Roles   []string   `json:"roles"`
					}
					if err := json.Unmarshal(payload.Data, &p); err == nil && p.GuildID != "" && p.User.ID != 0 {
						if mc, ok := c.Cache.(cache.MemberCache); ok {
							if cached, found := mc.GetMember(p.GuildID, p.User.ID.String()); found {
								if member, ok := cached.(*users.Member); ok {
									newMember := *member

									if newMember.User != nil {
										newUser := *newMember.User
										if p.User.Username != "" {
											newUser.Username = p.User.Username
										}
										if p.User.Discriminator != "" {
											newUser.Discriminator = p.User.Discriminator
										}
										if p.User.Avatar != nil {
											newUser.Avatar = p.User.Avatar
										}
										newMember.User = &newUser
									}

									if p.Roles != nil {
										newRoles := make([]snowflake.ID, len(p.Roles))
										for i, r := range p.Roles {
											if id, err := snowflake.Parse(r); err == nil {
												newRoles[i] = id
											}
										}
										newMember.Roles = newRoles
									}

									mc.SetMember(p.GuildID, p.User.ID.String(), &newMember)
								}
							}
						}
					}
				}
			}

			// Dispatch the raw data to handlers.
			c.Dispatcher.Dispatch(data)

		case OpcodeHeartbeat:
			// Gateway is requesting an immediate heartbeat.
			if c.Heartbeater != nil {
				if err := c.Heartbeater.sendHeartbeat(); err != nil {
					return err
				}
			}

		case OpcodeReconnect:
			// Gateway wants us to reconnect.
			c.Dispatcher.Dispatch(data)
			return ErrReconnectRequested

		case OpcodeInvalidSession:
			c.Dispatcher.Dispatch(data)
			return ErrInvalidSession

		case OpcodeHello:
			var hello HelloData
			if err := json.Unmarshal(payload.Data, &hello); err != nil {
				return fmt.Errorf("gateway: decode hello: %w", err)
			} else {
				if c.Heartbeater != nil {
					c.Heartbeater.Stop()
				}
				c.Heartbeater = NewHeartbeater(c.Conn, time.Duration(hello.HeartbeatInterval)*time.Millisecond)

				go func(hb *Heartbeater, conn Connection) {
					err := hb.Run(ctx)
					if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
						// Forcefully close the connection to trigger a reconnect
						_ = conn.Close()
					}
				}(c.Heartbeater, c.Conn)

				// Automatically Identify or Resume after receiving HELLO
				if c.Session != nil && c.Session.CanResume() {
					if err := c.sendResume(); err != nil {
						return err
					}
				} else {
					if c.IdentifyTracker != nil {
						if err := c.IdentifyTracker.Wait(ctx); err != nil {
							return err
						}
					}
					if err := c.sendIdentify(); err != nil {
						return err
					}
				}
			}
			c.Dispatcher.Dispatch(data)

		case OpcodeHeartbeatACK:
			if c.Heartbeater != nil {
				c.Heartbeater.AckReceived()
			}

		default:
			// Unknown opcode — dispatch the raw data.
			c.Dispatcher.Dispatch(data)
		}
	}
}

// sendIdentify sends an Identify (op 2) payload to the gateway.
func (c *Client) sendIdentify() error {
	identify := Identify{
		Token:   c.token,
		Intents: c.Intents,
		Shard:   c.Shard,
		Properties: IdentifyProperties{
			OS:      "linux",
			Browser: "discord.go",
			Device:  "discord.go",
		},
	}

	data, err := json.Marshal(identify)
	if err != nil {
		return err
	}

	payload := GatewayPayload{
		Op:   OpcodeIdentify,
		Data: data,
	}

	return c.Send(context.Background(), payload)
}

// sendResume sends a Resume (op 6) payload to the gateway.
func (c *Client) sendResume() error {
	if c.Session == nil {
		return ErrInvalidSession
	}

	resume := c.Session.ToResume(c.token)

	data, err := json.Marshal(resume)
	if err != nil {
		return err
	}

	payload := GatewayPayload{
		Op:   OpcodeResume,
		Data: data,
	}

	return c.Send(context.Background(), payload)
}

// extractCloseCode attempts to extract a websocket close code from an error.
func extractCloseCode(err error) (CloseCode, bool) {
	if err == nil {
		return 0, false
	}
	val := reflect.ValueOf(err)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		field := val.FieldByName("Code")
		if field.IsValid() && field.CanInt() {
			return CloseCode(field.Int()), true
		}
	}
	return 0, false
}
