package events

import (
	"encoding/json"
	"testing"
)

func TestGuildEvents(t *testing.T) {
	data := []byte(`{"id":"123"}`)

	var gc GuildCreate
	if err := json.Unmarshal(data, &gc); err != nil {
		t.Fatal(err)
	}

	var gu GuildUpdate
	if err := json.Unmarshal(data, &gu); err != nil {
		t.Fatal(err)
	}

	var gd GuildDelete
	if err := json.Unmarshal(data, &gd); err != nil {
		t.Fatal(err)
	}
}

func TestMessageEvents(t *testing.T) {
	data := []byte(`{"id":"123","channel_id":"456","guild_id":"789","user_id":"111","message_id":"222","emoji":{"id":"333"}}`)

	var mc MessageCreate
	if err := json.Unmarshal(data, &mc); err != nil {
		t.Fatal(err)
	}

	var mu MessageUpdate
	if err := json.Unmarshal(data, &mu); err != nil {
		t.Fatal(err)
	}

	var md MessageDelete
	if err := json.Unmarshal(data, &md); err != nil {
		t.Fatal(err)
	}

	var mra MessageReactionAdd
	if err := json.Unmarshal(data, &mra); err != nil {
		t.Fatal(err)
	}
}

func TestChannelEvents(t *testing.T) {
	data := []byte(`{"id":"123"}`)

	var cc ChannelCreate
	if err := json.Unmarshal(data, &cc); err != nil {
		t.Fatal(err)
	}

	var cu ChannelUpdate
	if err := json.Unmarshal(data, &cu); err != nil {
		t.Fatal(err)
	}
}

func TestUserEvents(t *testing.T) {
	data := []byte(`{"v":10,"user":{"id":"123"},"session_id":"abc"}`)

	var r Ready
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatal(err)
	}
}

func TestInteractionEvents(t *testing.T) {
	data := []byte(`{"id":"123"}`)

	var ic InteractionCreate
	if err := json.Unmarshal(data, &ic); err != nil {
		t.Fatal(err)
	}
}

func TestPayloads(t *testing.T) {
	data := []byte(`{"op":0,"d":{"foo":"bar"},"s":1,"t":"READY"}`)
	var e Event
	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatal(err)
	}
}
