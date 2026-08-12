package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestGetWebhook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"hook"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	wh, err := c.GetWebhook(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if wh.ID.String() != "123" {
		t.Errorf("Expected id 123, got %s", wh.ID.String())
	}
}

func TestModifyWebhook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"newhook"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	wh, err := c.ModifyWebhook(context.Background(), id, ModifyWebhookParams{Name: "newhook"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if wh.Name != "newhook" {
		t.Errorf("Expected name newhook, got %s", wh.Name)
	}
}

func TestDeleteWebhook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/123" {
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

	id, _ := snowflake.Parse("123")
	err := c.DeleteWebhook(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestExecuteWebhookParamsBuilder(t *testing.T) {
	b := NewExecuteWebhookParamsBuilder().
		SetContent("hello").
		SetUsername("Bot").
		SetTTS(false).
		SetFlags(64)

	params := b.Build()
	if params.Content != "hello" {
		t.Errorf("Content = %q, want %q", params.Content, "hello")
	}
	if params.Username != "Bot" {
		t.Errorf("Username = %q, want %q", params.Username, "Bot")
	}
	if params.Flags != 64 {
		t.Errorf("Flags = %d, want 64", params.Flags)
	}
}

func TestExecuteWebhook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/123/token" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("wait") != "true" {
			t.Errorf("Expected wait=true, got %s", r.URL.Query().Get("wait"))
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
	msg, err := c.ExecuteWebhook(context.Background(), id, "token", ExecuteWebhookParams{Content: "hello"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "456" {
		t.Errorf("Expected message id 456, got %s", msg.ID.String())
	}
}

func TestEditWebhookMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/123/token/messages/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","content":"edited"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	whID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	content := "edited"
	msg, err := c.EditWebhookMessage(context.Background(), whID, "token", msgID, EditMessageParams{Content: &content})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "456" {
		t.Errorf("Expected message id 456, got %s", msg.ID.String())
	}
}

func TestDeleteWebhookMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/webhooks/123/token/messages/456" {
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

	whID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.DeleteWebhookMessage(context.Background(), whID, "token", msgID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
