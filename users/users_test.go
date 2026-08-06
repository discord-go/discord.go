package users_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discord-go/discord.go/permissions"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

func TestUser_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"id": "80351110224678912",
		"username": "Nelly",
		"discriminator": "1337",
		"avatar": "8342729096ea3675442027bfc8ce3706",
		"verified": true,
		"email": "nelly@discord.com",
		"flags": 64,
		"premium_type": 1,
		"public_flags": 64
	}`)

	var u users.User
	err := json.Unmarshal(data, &u)
	if err != nil {
		t.Fatalf("Failed to unmarshal User: %v", err)
	}

	if u.ID != snowflake.ID(80351110224678912) {
		t.Errorf("Expected ID 80351110224678912, got %v", u.ID)
	}
	if u.Username != "Nelly" {
		t.Errorf("Expected Username Nelly, got %v", u.Username)
	}
	if u.Discriminator != "1337" {
		t.Errorf("Expected Discriminator 1337, got %v", u.Discriminator)
	}
	if u.Avatar == nil || *u.Avatar != "8342729096ea3675442027bfc8ce3706" {
		t.Errorf("Expected Avatar 8342729096ea3675442027bfc8ce3706, got %v", u.Avatar)
	}
	if !u.Verified {
		t.Errorf("Expected Verified true, got %v", u.Verified)
	}
	if u.Email == nil || *u.Email != "nelly@discord.com" {
		t.Errorf("Expected Email nelly@discord.com, got %v", u.Email)
	}
	if u.Flags != users.FlagHypesquadOnlineHouse1 {
		t.Errorf("Expected Flags 64, got %v", u.Flags)
	}
	if u.PremiumType != users.PremiumTypeNitroClassic {
		t.Errorf("Expected PremiumType 1, got %v", u.PremiumType)
	}
	if u.PublicFlags != users.FlagHypesquadOnlineHouse1 {
		t.Errorf("Expected PublicFlags 64, got %v", u.PublicFlags)
	}
}

func TestMember_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"user": {
			"id": "80351110224678912",
			"username": "Nelly",
			"discriminator": "1337",
			"avatar": "8342729096ea3675442027bfc8ce3706"
		},
		"nick": "Nelly!",
		"avatar": "1234567890",
		"roles": ["2233445566778899"],
		"joined_at": "2015-04-26T06:26:56.936000+00:00",
		"deaf": false,
		"mute": false,
		"permissions": "2147483647"
	}`)

	var m users.Member
	err := json.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("Failed to unmarshal Member: %v", err)
	}

	if m.User == nil || m.User.ID != snowflake.ID(80351110224678912) {
		t.Errorf("Expected User ID 80351110224678912, got %v", m.User)
	}
	if m.Nick == nil || *m.Nick != "Nelly!" {
		t.Errorf("Expected Nick Nelly!, got %v", m.Nick)
	}
	if m.Avatar == nil || *m.Avatar != "1234567890" {
		t.Errorf("Expected Avatar 1234567890, got %v", m.Avatar)
	}
	if len(m.Roles) != 1 || m.Roles[0] != snowflake.ID(2233445566778899) {
		t.Errorf("Expected Roles [2233445566778899], got %v", m.Roles)
	}
	expectedTime, _ := time.Parse(time.RFC3339Nano, "2015-04-26T06:26:56.936000+00:00")
	if !m.JoinedAt.Equal(expectedTime) {
		t.Errorf("Expected JoinedAt %v, got %v", expectedTime, m.JoinedAt)
	}
	if m.Deaf != false {
		t.Errorf("Expected Deaf false, got %v", m.Deaf)
	}
	if m.Mute != false {
		t.Errorf("Expected Mute false, got %v", m.Mute)
	}
	if m.Permissions != permissions.Permission(2147483647) {
		t.Errorf("Expected Permissions 2147483647, got %v", m.Permissions)
	}
}

func TestPresenceUpdate_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"user": {
			"id": "80351110224678912",
			"username": "Nelly",
			"discriminator": "1337"
		},
		"guild_id": "123456789",
		"status": "online",
		"activities": [
			{
				"name": "Rocket League",
				"type": 0,
				"created_at": 1501234567,
				"application_id": "123456"
			}
		],
		"client_status": {
			"desktop": "online",
			"mobile": "idle"
		}
	}`)

	var pu users.PresenceUpdate
	err := json.Unmarshal(data, &pu)
	if err != nil {
		t.Fatalf("Failed to unmarshal PresenceUpdate: %v", err)
	}

	if pu.User.ID != snowflake.ID(80351110224678912) {
		t.Errorf("Expected User ID 80351110224678912, got %v", pu.User.ID)
	}
	if pu.GuildID != snowflake.ID(123456789) {
		t.Errorf("Expected GuildID 123456789, got %v", pu.GuildID)
	}
	if pu.Status != "online" {
		t.Errorf("Expected Status online, got %v", pu.Status)
	}
	if len(pu.Activities) != 1 || pu.Activities[0].Name != "Rocket League" || pu.Activities[0].ApplicationID != snowflake.ID(123456) {
		t.Errorf("Unexpected Activities: %+v", pu.Activities)
	}
	if pu.ClientStatus.Desktop != "online" || pu.ClientStatus.Mobile != "idle" || pu.ClientStatus.Web != "" {
		t.Errorf("Unexpected ClientStatus: %+v", pu.ClientStatus)
	}
}

func TestFlagsAndPremium(t *testing.T) {
	if users.FlagStaff != 1<<0 {
		t.Errorf("Expected FlagStaff 1, got %v", users.FlagStaff)
	}
	if users.PremiumTypeNitroClassic != 1 {
		t.Errorf("Expected PremiumTypeNitroClassic 1, got %v", users.PremiumTypeNitroClassic)
	}
}

func TestMember_UnmarshalJSON_Error(t *testing.T) {
	// test malformed json
	var m users.Member
	if err := m.UnmarshalJSON([]byte(`{malformed`)); err == nil {
		t.Error("Expected error on malformed json, got nil")
	}

	// test invalid snowflake parse
	if err := json.Unmarshal([]byte(`{"roles": ["invalid"]}`), &m); err == nil {
		t.Error("Expected error on invalid snowflake, got nil")
	}

	// test no roles
	var m2 users.Member
	if err := json.Unmarshal([]byte(`{}`), &m2); err != nil {
		t.Error("Unexpected error when roles is omitted")
	}
}
