package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestListAutoModerationRules(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/auto-moderation/rules" {
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
	rules, err := c.ListAutoModerationRules(context.Background(), guildID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(rules) != 1 || rules[0].ID.String() != "456" {
		t.Errorf("Expected rule with id 456, got %v", rules)
	}
}

func TestGetAutoModerationRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/auto-moderation/rules/456" {
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
	ruleID, _ := snowflake.Parse("456")
	rule, err := c.GetAutoModerationRule(context.Background(), guildID, ruleID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if rule.ID.String() != "456" {
		t.Errorf("Expected rule id 456, got %s", rule.ID.String())
	}
}

func TestCreateAutoModerationRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/auto-moderation/rules" {
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
	params := CreateAutoModerationRuleParams{Name: "test"}
	rule, err := c.CreateAutoModerationRule(context.Background(), guildID, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if rule.ID.String() != "456" {
		t.Errorf("Expected rule id 456, got %s", rule.ID.String())
	}
}

func TestModifyAutoModerationRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/auto-moderation/rules/456" {
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
	ruleID, _ := snowflake.Parse("456")
	name := "new name"
	params := ModifyAutoModerationRuleParams{Name: &name}
	rule, err := c.ModifyAutoModerationRule(context.Background(), guildID, ruleID, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if rule.ID.String() != "456" {
		t.Errorf("Expected rule id 456, got %s", rule.ID.String())
	}
}

func TestDeleteAutoModerationRule(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/auto-moderation/rules/456" {
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
	ruleID, _ := snowflake.Parse("456")
	err := c.DeleteAutoModerationRule(context.Background(), guildID, ruleID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
