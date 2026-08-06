package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestCreateMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","content":"hello"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	msg, err := c.CreateMessage(context.Background(), id, "hello")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "456" {
		t.Errorf("Expected message id 456, got %s", msg.ID.String())
	}
	if msg.Content != "hello" {
		t.Errorf("Expected content hello, got %s", msg.Content)
	}
}

func TestCreateMessage_Error(t *testing.T) {
	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	})
	id, _ := snowflake.Parse("123")
	_, err := c.CreateMessage(context.Background(), id, "hello")
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetGuild(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/789" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"789","name":"test guild"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("789")
	guild, err := c.GetGuild(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if guild.ID.String() != "789" {
		t.Errorf("Expected guild id 789, got %s", guild.ID.String())
	}
	if guild.Name != "test guild" {
		t.Errorf("Expected name 'test guild', got %s", guild.Name)
	}
}

func TestGetGuild_Error(t *testing.T) {
	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	})
	id, _ := snowflake.Parse("789")
	_, err := c.GetGuild(context.Background(), id)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}
