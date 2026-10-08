package gateway

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestOpcodeRequestChannelInfoValue(t *testing.T) {
	if OpcodeRequestChannelInfo != 43 {
		t.Errorf("OpcodeRequestChannelInfo = %d, want 43", OpcodeRequestChannelInfo)
	}
}

func TestRequestChannelInfoSendsOpcode43(t *testing.T) {
	conn := &captureConnection{}
	c := NewClient(conn, NewDispatcher())

	data := RequestChannelInfoData{
		GuildID: snowflake.ID(613425648685547541),
		Fields:  []string{ChannelInfoFieldStatus, ChannelInfoFieldVoiceStartTime},
	}
	if err := c.RequestChannelInfoContext(context.Background(), data); err != nil {
		t.Fatalf("RequestChannelInfoContext: %v", err)
	}
	if len(conn.writes) != 1 {
		t.Fatalf("writes = %d, want 1", len(conn.writes))
	}

	var payload GatewayPayload
	if err := json.Unmarshal(conn.writes[0], &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Op != OpcodeRequestChannelInfo {
		t.Errorf("op = %d, want %d", payload.Op, OpcodeRequestChannelInfo)
	}
	var sent RequestChannelInfoData
	if err := json.Unmarshal(payload.Data, &sent); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if sent.GuildID != snowflake.ID(613425648685547541) {
		t.Errorf("guild_id = %s", sent.GuildID)
	}
	if len(sent.Fields) != 2 || sent.Fields[0] != ChannelInfoFieldStatus || sent.Fields[1] != ChannelInfoFieldVoiceStartTime {
		t.Errorf("fields = %v", sent.Fields)
	}
}

func TestRequestChannelInfoFieldsConstants(t *testing.T) {
	if ChannelInfoFieldStatus != "status" {
		t.Errorf("ChannelInfoFieldStatus = %q", ChannelInfoFieldStatus)
	}
	if ChannelInfoFieldVoiceStartTime != "voice_start_time" {
		t.Errorf("ChannelInfoFieldVoiceStartTime = %q", ChannelInfoFieldVoiceStartTime)
	}
}
