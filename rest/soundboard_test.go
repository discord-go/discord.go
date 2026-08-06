package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestListDefaultSoundboardSounds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/soundboard-default-sounds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"sound_id":"456","name":"sound1"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	sounds, err := c.ListDefaultSoundboardSounds(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(sounds) != 1 || sounds[0].SoundID.String() != "456" {
		t.Errorf("Expected 1 sound with id 456, got %v", sounds)
	}
}

func TestListGuildSoundboardSounds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/soundboard-sounds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"sound_id":"456","name":"sound1"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	sounds, err := c.ListGuildSoundboardSounds(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(sounds) != 1 || sounds[0].SoundID.String() != "456" {
		t.Errorf("Expected 1 sound with id 456, got %v", sounds)
	}
}

func TestGetGuildSoundboardSound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/soundboard-sounds/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sound_id":"456","name":"sound1"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	soundID, _ := snowflake.Parse("456")
	sound, err := c.GetGuildSoundboardSound(context.Background(), guildID, soundID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sound.SoundID.String() != "456" {
		t.Errorf("Expected sound id 456, got %s", sound.SoundID.String())
	}
}

func TestCreateGuildSoundboardSound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/soundboard-sounds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sound_id":"456","name":"new_sound"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	sound, err := c.CreateGuildSoundboardSound(context.Background(), guildID, CreateSoundboardSoundParams{Name: "new_sound"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sound.Name != "new_sound" {
		t.Errorf("Expected name new_sound, got %s", sound.Name)
	}
}

func TestModifyGuildSoundboardSound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/soundboard-sounds/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sound_id":"456","name":"modified_sound"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	soundID, _ := snowflake.Parse("456")
	newName := "modified_sound"
	sound, err := c.ModifyGuildSoundboardSound(context.Background(), guildID, soundID, ModifySoundboardSoundParams{Name: &newName})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sound.Name != "modified_sound" {
		t.Errorf("Expected name modified_sound, got %s", sound.Name)
	}
}

func TestDeleteGuildSoundboardSound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/soundboard-sounds/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	soundID, _ := snowflake.Parse("456")
	err := c.DeleteGuildSoundboardSound(context.Background(), guildID, soundID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
