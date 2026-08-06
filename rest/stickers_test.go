package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestListGuildStickers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/stickers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456","name":"sticker1"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	stickers, err := c.ListGuildStickers(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(stickers) != 1 || stickers[0].ID.String() != "456" {
		t.Errorf("Expected 1 sticker with id 456, got %v", stickers)
	}
}

func TestGetGuildSticker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/stickers/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","name":"sticker1"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	stickerID, _ := snowflake.Parse("456")
	sticker, err := c.GetGuildSticker(context.Background(), guildID, stickerID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sticker.ID.String() != "456" {
		t.Errorf("Expected sticker id 456, got %s", sticker.ID.String())
	}
}

func TestCreateGuildSticker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/stickers" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","name":"new_sticker"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	sticker, err := c.CreateGuildSticker(context.Background(), guildID, CreateStickerParams{Name: "new_sticker"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sticker.Name != "new_sticker" {
		t.Errorf("Expected name new_sticker, got %s", sticker.Name)
	}
}

func TestModifyGuildSticker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/stickers/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","name":"modified_sticker"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	stickerID, _ := snowflake.Parse("456")
	newName := "modified_sticker"
	sticker, err := c.ModifyGuildSticker(context.Background(), guildID, stickerID, ModifyStickerParams{Name: &newName})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if sticker.Name != "modified_sticker" {
		t.Errorf("Expected name modified_sticker, got %s", sticker.Name)
	}
}

func TestDeleteGuildSticker(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/stickers/456" {
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
	stickerID, _ := snowflake.Parse("456")
	err := c.DeleteGuildSticker(context.Background(), guildID, stickerID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
