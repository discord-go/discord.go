package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentApplicationStringFlagsAndWebhookEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/applications/@me" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":"42","name":"app","flags_new":"8192","event_webhooks_types":["APPLICATION_AUTHORIZED"]}`)
	}))
	defer server.Close()
	c := New("token", nil, server.Client())
	c.BaseURL = server.URL
	app, err := c.GetCurrentApplication(context.Background())
	if err != nil || app.FlagsNew != 8192 || len(app.EventWebhookTypeNames) != 1 || app.EventWebhookTypeNames[0] != "APPLICATION_AUTHORIZED" {
		t.Fatalf("application = %+v, error = %v", app, err)
	}
}
