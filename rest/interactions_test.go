package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/snowflake"
)

func TestCreateInteractionResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/interactions/123/token/callback" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
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

	id, _ := snowflake.Parse("123")
	err := c.CreateInteractionResponse(context.Background(), id, "token", interactions.InteractionResponse{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGetOriginalInteractionResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/456/token/messages/@original" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"789","content":"hello"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("456")
	msg, err := c.GetOriginalInteractionResponse(context.Background(), appID, "token")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "789" {
		t.Errorf("Expected message id 789, got %s", msg.ID.String())
	}
}

func TestEditOriginalInteractionResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/456/token/messages/@original" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"789","content":"edited"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("456")
	content := "edited"
	msg, err := c.EditOriginalInteractionResponse(context.Background(), appID, "token", EditMessageParams{Content: &content})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.Content != "edited" {
		t.Errorf("Expected edited, got %s", msg.Content)
	}
}

func TestDeleteOriginalInteractionResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/456/token/messages/@original" {
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

	appID, _ := snowflake.Parse("456")
	err := c.DeleteOriginalInteractionResponse(context.Background(), appID, "token")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGetGlobalApplicationCommands(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/456/commands" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"111","name":"cmd"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("456")
	cmds, err := c.GetGlobalApplicationCommands(context.Background(), appID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cmds) != 1 || cmds[0].ID.String() != "111" {
		t.Errorf("Expected 1 cmd with id 111, got %v", cmds)
	}
}

func TestCreateGlobalApplicationCommand(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/456/commands" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"111","name":"cmd"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("456")
	cmd, err := c.CreateGlobalApplicationCommand(context.Background(), appID, CreateCommandParams{Name: "cmd"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cmd.ID.String() != "111" {
		t.Errorf("Expected cmd id 111, got %s", cmd.ID.String())
	}
}

func TestDeleteGlobalApplicationCommand(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/456/commands/111" {
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

	appID, _ := snowflake.Parse("456")
	cmdID, _ := snowflake.Parse("111")
	err := c.DeleteGlobalApplicationCommand(context.Background(), appID, cmdID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestBulkOverwriteGlobalCommands(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/applications/456/commands" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"111","name":"cmd"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	appID, _ := snowflake.Parse("456")
	cmds, err := c.BulkOverwriteGlobalCommands(context.Background(), appID, []CreateCommandParams{{Name: "cmd"}})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(cmds) != 1 || cmds[0].ID.String() != "111" {
		t.Errorf("Expected 1 cmd with id 111, got %v", cmds)
	}
}
