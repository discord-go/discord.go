package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

func TestFormatTranscript(t *testing.T) {
	globalName := "Alice"
	msgs := []messages.Message{
		{
			ID:        snowflake.ID(100),
			Content:   "Hello, world!",
			Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			Author: &users.User{
				ID:         snowflake.ID(1),
				Username:   "alice",
				GlobalName: &globalName,
			},
		},
		{
			ID:        snowflake.ID(101),
			Content:   "Check this file",
			Timestamp: time.Date(2025, 1, 15, 10, 31, 0, 0, time.UTC),
			Author: &users.User{
				ID:       snowflake.ID(2),
				Username: "bob",
			},
			Attachments: []messages.Attachment{
				{Filename: "report.pdf", URL: "https://example.com/report.pdf"},
			},
			Embeds: []messages.Embed{{Title: "Summary"}},
		},
	}

	result := FormatTranscript(msgs, TranscriptOptions{
		ChannelName: "ticket-001",
		ChannelID:   snowflake.ID(999),
		GeneratedAt: time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC),
	})

	if !strings.Contains(result, "Transcript for #ticket-001") {
		t.Error("expected channel name in transcript header")
	}
	if !strings.Contains(result, "Channel ID: 999") {
		t.Error("expected channel ID in transcript header")
	}
	if !strings.Contains(result, "Messages: 2") {
		t.Error("expected message count in transcript header")
	}
	if !strings.Contains(result, "Alice (1): Hello, world!") {
		t.Error("expected Alice's message with global name")
	}
	if !strings.Contains(result, "bob (2): Check this file") {
		t.Error("expected bob's message with username")
	}
	if !strings.Contains(result, "Attachment: report.pdf") {
		t.Error("expected attachment in transcript")
	}
	if !strings.Contains(result, "Embed: Summary") {
		t.Error("expected embed in transcript")
	}
}

func TestFormatTranscriptEmpty(t *testing.T) {
	result := FormatTranscript(nil, TranscriptOptions{
		ChannelName: "empty-channel",
		ChannelID:   snowflake.ID(42),
	})

	if !strings.Contains(result, "Messages: 0") {
		t.Error("expected zero messages in transcript header")
	}
	if !strings.Contains(result, "Transcript for #empty-channel") {
		t.Error("expected channel name in header")
	}
}

func TestFormatTranscriptNilAuthor(t *testing.T) {
	msgs := []messages.Message{
		{
			ID:        snowflake.ID(100),
			Content:   "Mystery message",
			Timestamp: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	result := FormatTranscript(msgs, TranscriptOptions{
		ChannelID: snowflake.ID(1),
	})

	if !strings.Contains(result, "Unknown (): Mystery message") {
		t.Error("expected Unknown author for nil author")
	}
}

func TestFetchAllMessages(t *testing.T) {
	// Simulate a channel with 150 messages across two pages.
	page1 := make([]messages.Message, 100)
	for i := 0; i < 100; i++ {
		page1[i] = messages.Message{
			ID:        snowflake.ID(200 - i), // 200, 199, ..., 101
			Content:   "msg " + string(rune('A'+i%26)),
			Timestamp: time.Now(),
		}
	}
	page2 := make([]messages.Message, 50)
	for i := 0; i < 50; i++ {
		page2[i] = messages.Message{
			ID:        snowflake.ID(100 - i), // 100, 99, ..., 51
			Content:   "msg " + string(rune('a'+i%26)),
			Timestamp: time.Now(),
		}
	}

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var page []messages.Message
		if callCount == 1 {
			page = page1
		} else {
			page = page2
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	msgs, err := c.FetchAllMessages(context.Background(), snowflake.ID(1), 0)
	if err != nil {
		t.Fatalf("FetchAllMessages error: %v", err)
	}
	if len(msgs) != 150 {
		t.Fatalf("expected 150 messages, got %d", len(msgs))
	}
	// Verify chronological order: first message should have the lowest ID.
	if msgs[0].ID != snowflake.ID(51) {
		t.Errorf("expected first message ID 51, got %s", msgs[0].ID)
	}
	if msgs[149].ID != snowflake.ID(200) {
		t.Errorf("expected last message ID 200, got %s", msgs[149].ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestFetchAllMessagesWithMax(t *testing.T) {
	// 100 messages available, but we cap at 25.
	allMsgs := make([]messages.Message, 100)
	for i := 0; i < 100; i++ {
		allMsgs[i] = messages.Message{
			ID:        snowflake.ID(300 - i),
			Content:   "capped",
			Timestamp: time.Now(),
		}
	}

	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Parse the limit query parameter.
		limit := 100
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := parseLimitParam(l); err == nil {
				limit = n
			}
		}
		page := allMsgs[:limit]
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	msgs, err := c.FetchAllMessages(context.Background(), snowflake.ID(1), 25)
	if err != nil {
		t.Fatalf("FetchAllMessages error: %v", err)
	}
	if len(msgs) != 25 {
		t.Fatalf("expected 25 messages, got %d", len(msgs))
	}
	if callCount != 1 {
		t.Errorf("expected 1 API call, got %d", callCount)
	}
}

func TestFetchAllMessagesEmpty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	msgs, err := c.FetchAllMessages(context.Background(), snowflake.ID(1), 0)
	if err != nil {
		t.Fatalf("FetchAllMessages error: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}
}

func TestGenerateTranscript(t *testing.T) {
	page := []messages.Message{
		{
			ID:        snowflake.ID(1),
			Content:   "Test transcript generation",
			Timestamp: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
			Author: &users.User{
				ID:       snowflake.ID(10),
				Username: "tester",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	file, err := c.GenerateTranscript(context.Background(), snowflake.ID(42), 0, "my-transcript.txt", TranscriptOptions{
		ChannelName: "ticket-42",
	})
	if err != nil {
		t.Fatalf("GenerateTranscript error: %v", err)
	}
	if file.Name != "my-transcript.txt" {
		t.Errorf("expected filename my-transcript.txt, got %s", file.Name)
	}

	content, err := io.ReadAll(file.Reader)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "Transcript for #ticket-42") {
		t.Error("expected channel name in generated transcript")
	}
	if !strings.Contains(s, "tester (10): Test transcript generation") {
		t.Error("expected message content in generated transcript")
	}
}

func TestGenerateTranscriptDefaultFilename(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	file, err := c.GenerateTranscript(context.Background(), snowflake.ID(1), 0, "", TranscriptOptions{})
	if err != nil {
		t.Fatalf("GenerateTranscript error: %v", err)
	}
	if file.Name != "transcript.txt" {
		t.Errorf("expected default filename transcript.txt, got %s", file.Name)
	}
}

func parseLimitParam(s string) (int, error) {
	return strconv.Atoi(s)
}
