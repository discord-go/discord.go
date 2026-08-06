package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestGetCurrentUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","username":"testuser"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	user, err := c.GetCurrentUser(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if user.ID.String() != "123" {
		t.Errorf("Expected id 123, got %s", user.ID.String())
	}
}

func TestGetUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","username":"otheruser"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("456")
	user, err := c.GetUser(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if user.ID.String() != "456" {
		t.Errorf("Expected id 456, got %s", user.ID.String())
	}
}

func TestModifyCurrentUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","username":"newname"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	user, err := c.ModifyCurrentUser(context.Background(), ModifyUserParams{Username: "newname"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if user.Username != "newname" {
		t.Errorf("Expected username newname, got %s", user.Username)
	}
}

func TestGetCurrentUserGuilds(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me/guilds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"789","name":"guild1"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	guilds, err := c.GetCurrentUserGuilds(context.Background(), ListGuildsParams{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(guilds) != 1 || guilds[0].ID.String() != "789" {
		t.Errorf("Expected 1 guild with id 789, got %v", guilds)
	}
}

func TestLeaveGuild(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me/guilds/789" {
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

	id, _ := snowflake.Parse("789")
	err := c.LeaveGuild(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestCreateDM(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/@me/channels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"999"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("456")
	ch, err := c.CreateDM(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ch.ID.String() != "999" {
		t.Errorf("Expected channel id 999, got %s", ch.ID.String())
	}
}
