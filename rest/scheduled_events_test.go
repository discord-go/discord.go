package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestListScheduledEvents(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/scheduled-events" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	events, err := c.ListScheduledEvents(context.Background(), guildID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(events) != 1 || events[0].ID.String() != "456" {
		t.Errorf("Expected event with id 456, got %v", events)
	}
}

func TestGetScheduledEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/scheduled-events/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	eventID, _ := snowflake.Parse("456")
	event, err := c.GetScheduledEvent(context.Background(), guildID, eventID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if event.ID.String() != "456" {
		t.Errorf("Expected event id 456, got %s", event.ID.String())
	}
}

func TestCreateScheduledEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/scheduled-events" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	params := CreateScheduledEventParams{Name: "test"}
	event, err := c.CreateScheduledEvent(context.Background(), guildID, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if event.ID.String() != "456" {
		t.Errorf("Expected event id 456, got %s", event.ID.String())
	}
}

func TestModifyScheduledEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/scheduled-events/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guildID, _ := snowflake.Parse("123")
	eventID, _ := snowflake.Parse("456")
	name := "new name"
	params := ModifyScheduledEventParams{Name: &name}
	event, err := c.ModifyScheduledEvent(context.Background(), guildID, eventID, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if event.ID.String() != "456" {
		t.Errorf("Expected event id 456, got %s", event.ID.String())
	}
}

func TestDeleteScheduledEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/scheduled-events/456" {
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
	eventID, _ := snowflake.Parse("456")
	err := c.DeleteScheduledEvent(context.Background(), guildID, eventID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
