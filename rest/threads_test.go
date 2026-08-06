package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestStartThreadWithMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456/threads" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"789"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	channelID, _ := snowflake.Parse("123")
	messageID, _ := snowflake.Parse("456")
	params := StartThreadWithMessageParams{Name: "test"}
	thread, err := c.StartThreadWithMessage(context.Background(), channelID, messageID, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if thread.ID.String() != "789" {
		t.Errorf("Expected thread id 789, got %s", thread.ID.String())
	}
}

func TestStartThread(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/threads" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"789"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	channelID, _ := snowflake.Parse("123")
	params := StartThreadParams{Name: "test"}
	thread, err := c.StartThread(context.Background(), channelID, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if thread.ID.String() != "789" {
		t.Errorf("Expected thread id 789, got %s", thread.ID.String())
	}
}

func TestJoinThread(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/thread-members/@me" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
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

	threadID, _ := snowflake.Parse("123")
	err := c.JoinThread(context.Background(), threadID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestAddThreadMember(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/thread-members/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
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

	threadID, _ := snowflake.Parse("123")
	userID, _ := snowflake.Parse("456")
	err := c.AddThreadMember(context.Background(), threadID, userID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestListActiveThreads(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/threads/active" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"threads":[{"id":"456"}],"members":[]}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	active, err := c.ListActiveThreads(context.Background(), guildID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(active.Threads) != 1 || active.Threads[0].ID.String() != "456" {
		t.Errorf("Expected 1 thread with id 456, got %v", active.Threads)
	}
}
