package channels_test

import (
	"encoding/json"
	"testing"

	"github.com/discord-go/discord.go/channels"
)

func TestChannelUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "123456789",
		"type": 0,
		"guild_id": "987654321",
		"name": "general",
		"permission_overwrites": [
			{
				"id": "111",
				"type": 0,
				"allow": "8",
				"deny": "0"
			}
		],
		"available_tags": [
			{
				"id": "222",
				"name": "tag1",
				"moderated": false
			}
		],
		"default_reaction_emoji": {
			"emoji_name": "smile"
		}
	}`)
	var c channels.Channel
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("Failed to unmarshal Channel: %v", err)
	}
}

func TestThreadMetadataUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"archived": true,
		"auto_archive_duration": 60,
		"archive_timestamp": "2021-01-01T00:00:00Z",
		"locked": false
	}`)
	var tm channels.ThreadMetadata
	if err := json.Unmarshal(data, &tm); err != nil {
		t.Fatalf("Failed to unmarshal ThreadMetadata: %v", err)
	}
}

func TestThreadMemberUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "123",
		"user_id": "456",
		"join_timestamp": "2021-01-01T00:00:00Z",
		"flags": 1
	}`)
	var tm channels.ThreadMember
	if err := json.Unmarshal(data, &tm); err != nil {
		t.Fatalf("Failed to unmarshal ThreadMember: %v", err)
	}
}

func TestInviteUnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"code": "abc",
		"guild": {
			"id": "123",
			"name": "my guild",
			"features": ["COMMUNITY"],
			"verification_level": 1
		},
		"channel": {
			"id": "456",
			"type": 0
		},
		"target_application": {
			"id": "789",
			"name": "app",
			"description": "desc"
		}
	}`)
	var inv channels.Invite
	if err := json.Unmarshal(data, &inv); err != nil {
		t.Fatalf("Failed to unmarshal Invite: %v", err)
	}
}
