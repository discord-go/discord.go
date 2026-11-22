package bot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/users"
)

type paginatorRequest struct {
	method, path string
	body         map[string]json.RawMessage
}

func paginatorTestBot(t *testing.T, callbackStatus, patchStatus int, patchGate ...<-chan struct{}) (*Bot, <-chan paginatorRequest) {
	t.Helper()
	requests := make(chan paginatorRequest, 30)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		requests <- paginatorRequest{r.Method, r.URL.Path, body}
		if r.Method == http.MethodPatch && len(patchGate) > 0 {
			<-patchGate[0]
		}
		if strings.Contains(r.URL.Path, "/callback") {
			if callbackStatus != 0 {
				w.WriteHeader(callbackStatus)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPatch && patchStatus != 0 {
			w.WriteHeader(patchStatus)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"42","channel_id":"10"}`))
	}))
	t.Cleanup(server.Close)
	b := New("token")
	b.Rest.BaseURL = server.URL
	return b, requests
}

func waitPaginatorCollector(t *testing.T, b *Bot) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		b.collectorMu.Lock()
		n := len(b.interactionCollectors)
		b.collectorMu.Unlock()
		if n > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("paginator collector was not registered")
}

func paginatorClick(b *Bot, id string, user *users.User) {
	data, _ := json.Marshal(map[string]string{"custom_id": id})
	ic := newInteractionContext(BaseContext{Bot: b, rest: b.Rest, ctx: context.Background()}, &interactions.Interaction{
		ID: 100, Type: interactions.InteractionTypeMessageComponent, Token: "test-token", Data: data,
		Message: &messages.Message{ID: 42}, User: user,
	})
	b.publishInteraction(ic)
}

func nextPaginatorRequest(t *testing.T, requests <-chan paginatorRequest) paginatorRequest {
	t.Helper()
	select {
	case request := <-requests:
		return request
	case <-time.After(2 * time.Second):
		t.Fatal("request timed out")
		return paginatorRequest{}
	}
}

func requireButtonsRemoved(t *testing.T, request paginatorRequest) {
	t.Helper()
	if request.method != http.MethodPatch || string(request.body["components"]) != "[]" {
		t.Fatalf("expected PATCH clearing components, got %s %s", request.method, request.body["components"])
	}
}

func TestPaginatorNavigationAndStop(t *testing.T) {
	b, requests := paginatorTestBot(t, 0, 0)
	p := b.NewPaginator(10, []PaginatorPage{{Content: "first"}, {Content: "second"}}, WithPaginatorUser(5))
	done := make(chan error, 1)
	go func() { done <- p.Send(context.Background()) }()
	nextPaginatorRequest(t, requests)
	waitPaginatorCollector(t, b)
	paginatorClick(b, paginatorNextID, nil)
	paginatorClick(b, paginatorNextID, &users.User{ID: 5})
	callback := nextPaginatorRequest(t, requests)
	if callback.method != http.MethodPost || string(callback.body["type"]) != "6" {
		t.Fatalf("callback: %+v", callback)
	}
	edit := nextPaginatorRequest(t, requests)
	if edit.method != http.MethodPatch || string(edit.body["content"]) != `"second"` {
		t.Fatalf("page edit: %+v", edit)
	}
	paginatorClick(b, paginatorStopID, &users.User{ID: 5})
	callback = nextPaginatorRequest(t, requests)
	if string(callback.body["type"]) != "6" {
		t.Fatalf("stop callback: %+v", callback)
	}
	requireButtonsRemoved(t, nextPaginatorRequest(t, requests))
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestPaginatorCleanupOnTimeoutAndCancellation(t *testing.T) {
	for _, cancelCaller := range []bool{false, true} {
		t.Run(map[bool]string{false: "timeout", true: "cancellation"}[cancelCaller], func(t *testing.T) {
			b, requests := paginatorTestBot(t, 0, 0)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := b.NewPaginator(10, []PaginatorPage{{Content: "first"}}, WithPaginatorTimeout(50*time.Millisecond))
			done := make(chan error, 1)
			go func() { done <- p.Send(ctx) }()
			nextPaginatorRequest(t, requests)
			waitPaginatorCollector(t, b)
			if cancelCaller {
				cancel()
			}
			requireButtonsRemoved(t, nextPaginatorRequest(t, requests))
			err := <-done
			if cancelCaller && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancel error: %v", err)
			}
			if !cancelCaller && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPaginatorReportsAcknowledgementAndCleanupErrors(t *testing.T) {
	b, requests := paginatorTestBot(t, http.StatusBadRequest, http.StatusBadRequest)
	p := b.NewPaginator(10, []PaginatorPage{{Content: "first"}})
	done := make(chan error, 1)
	go func() { done <- p.Send(context.Background()) }()
	nextPaginatorRequest(t, requests)
	waitPaginatorCollector(t, b)
	paginatorClick(b, paginatorNextID, &users.User{ID: 5})
	nextPaginatorRequest(t, requests)
	requireButtonsRemoved(t, nextPaginatorRequest(t, requests))
	err := <-done
	if err == nil || !strings.Contains(err.Error(), "acknowledge") || !strings.Contains(err.Error(), "remove buttons") {
		t.Fatalf("expected both errors: %v", err)
	}
}

// A second click arriving during the first REST edit must stay collected.
func TestPaginatorCollectsWhileEditing(t *testing.T) {
	gate := make(chan struct{})
	b, requests := paginatorTestBot(t, 0, 0, gate)
	p := b.NewPaginator(10, []PaginatorPage{{Content: "first"}, {Content: "second"}})
	done := make(chan error, 1)
	go func() { done <- p.Send(context.Background()) }()
	nextPaginatorRequest(t, requests)
	waitPaginatorCollector(t, b)
	paginatorClick(b, paginatorNextID, &users.User{ID: 5})
	nextPaginatorRequest(t, requests) // acknowledgement
	nextPaginatorRequest(t, requests) // edit is now waiting at the server
	paginatorClick(b, paginatorStopID, &users.User{ID: 5})
	close(gate)
	callback := nextPaginatorRequest(t, requests)
	if string(callback.body["type"]) != "6" {
		t.Fatalf("stop callback: %+v", callback)
	}
	requireButtonsRemoved(t, nextPaginatorRequest(t, requests))
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
