package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func setupTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	ts := httptest.NewServer(handler)
	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL
	return ts, c
}

func setupErrorClient() *Client {
	return New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, io.EOF
		},
	})
}

func TestGetChannel(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"general"}`))
	})
	defer ts.Close()

	id, _ := snowflake.Parse("123")
	ch, err := c.GetChannel(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ch.ID.String() != "123" {
		t.Errorf("Expected id 123, got %s", ch.ID.String())
	}
	if ch.Name == nil || *ch.Name != "general" {
		t.Errorf("Expected name 'general'")
	}
}

func TestGetChannel_Error(t *testing.T) {
	c := setupErrorClient()
	id, _ := snowflake.Parse("123")
	_, err := c.GetChannel(context.Background(), id)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestModifyChannel(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"new-name"}`))
	})
	defer ts.Close()

	id, _ := snowflake.Parse("123")
	name := "new-name"
	params := ModifyChannelParams{Name: &name}
	ch, err := c.ModifyChannel(context.Background(), id, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ch.ID.String() != "123" {
		t.Errorf("Expected id 123, got %s", ch.ID.String())
	}
	if ch.Name == nil || *ch.Name != "new-name" {
		t.Errorf("Expected name 'new-name'")
	}
}

func TestModifyChannel_Error(t *testing.T) {
	c := setupErrorClient()
	id, _ := snowflake.Parse("123")
	_, err := c.ModifyChannel(context.Background(), id, ModifyChannelParams{})
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestDeleteChannel(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	id, _ := snowflake.Parse("123")
	err := c.DeleteChannel(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDeleteChannel_Error(t *testing.T) {
	c := setupErrorClient()
	id, _ := snowflake.Parse("123")
	err := c.DeleteChannel(context.Background(), id)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetChannelMessages(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		if r.URL.RawQuery != "limit=50" {
			t.Errorf("Unexpected query: %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456","content":"msg1"}]`))
	})
	defer ts.Close()

	id, _ := snowflake.Parse("123")
	limit := 50
	msgs, err := c.GetChannelMessages(context.Background(), id, GetMessagesParams{Limit: &limit})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(msgs) != 1 || msgs[0].ID.String() != "456" {
		t.Errorf("Unexpected messages: %+v", msgs)
	}
}

func TestGetChannelMessages_Error(t *testing.T) {
	c := setupErrorClient()
	id, _ := snowflake.Parse("123")
	_, err := c.GetChannelMessages(context.Background(), id, GetMessagesParams{})
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetChannelMessage(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","content":"msg1"}`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	msg, err := c.GetChannelMessage(context.Background(), chID, msgID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "456" {
		t.Errorf("Expected msg id 456, got %s", msg.ID.String())
	}
}

func TestGetChannelMessage_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	_, err := c.GetChannelMessage(context.Background(), chID, msgID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestEditMessage(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","content":"updated"}`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	content := "updated"
	msg, err := c.EditMessage(context.Background(), chID, msgID, EditMessageParams{Content: &content})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "456" || msg.Content != "updated" {
		t.Errorf("Unexpected message: %+v", msg)
	}
}

func TestEditMessage_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	_, err := c.EditMessage(context.Background(), chID, msgID, EditMessageParams{})
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestDeleteMessage(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.DeleteMessage(context.Background(), chID, msgID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDeleteMessage_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.DeleteMessage(context.Background(), chID, msgID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestBulkDeleteMessages(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/bulk-delete" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msg1, _ := snowflake.Parse("456")
	msg2, _ := snowflake.Parse("789")
	err := c.BulkDeleteMessages(context.Background(), chID, []snowflake.ID{msg1, msg2})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestBulkDeleteMessages_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msg1, _ := snowflake.Parse("456")
	err := c.BulkDeleteMessages(context.Background(), chID, []snowflake.ID{msg1})
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestCreateReaction(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456/reactions/👍/@me" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.CreateReaction(context.Background(), chID, msgID, "👍")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestCreateReaction_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.CreateReaction(context.Background(), chID, msgID, "👍")
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestDeleteOwnReaction(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456/reactions/👍/@me" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.DeleteOwnReaction(context.Background(), chID, msgID, "👍")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestDeleteOwnReaction_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.DeleteOwnReaction(context.Background(), chID, msgID, "👍")
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetReactions(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/456/reactions/👍" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"789","username":"testuser"}]`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	users, err := c.GetReactions(context.Background(), chID, msgID, "👍")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(users) != 1 || users[0].ID.String() != "789" {
		t.Errorf("Unexpected users: %+v", users)
	}
}

func TestGetReactions_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	_, err := c.GetReactions(context.Background(), chID, msgID, "👍")
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestTriggerTypingIndicator(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/typing" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	err := c.TriggerTypingIndicator(context.Background(), chID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestTriggerTypingIndicator_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	err := c.TriggerTypingIndicator(context.Background(), chID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetPinnedMessages(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/pins" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456","content":"pinned"}]`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgs, err := c.GetPinnedMessages(context.Background(), chID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(msgs) != 1 || msgs[0].ID.String() != "456" {
		t.Errorf("Unexpected pinned msgs: %+v", msgs)
	}
}

func TestGetPinnedMessages_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	_, err := c.GetPinnedMessages(context.Background(), chID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestPinMessage(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/pins/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.PinMessage(context.Background(), chID, msgID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestPinMessage_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.PinMessage(context.Background(), chID, msgID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestUnpinMessage(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/messages/pins/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "DELETE" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.UnpinMessage(context.Background(), chID, msgID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestUnpinMessage_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	err := c.UnpinMessage(context.Background(), chID, msgID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetChannelInvites(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/invites" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"code":"abc"}]`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	invites, err := c.GetChannelInvites(context.Background(), chID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(invites) != 1 || invites[0].Code != "abc" {
		t.Errorf("Unexpected invites: %+v", invites)
	}
}

func TestGetChannelInvites_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	_, err := c.GetChannelInvites(context.Background(), chID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestCreateChannelInvite(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/invites" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"def"}`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	invite, err := c.CreateChannelInvite(context.Background(), chID, CreateInviteParams{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if invite.Code != "def" {
		t.Errorf("Expected invite code def, got %s", invite.Code)
	}
}

func TestCreateChannelInvite_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	_, err := c.CreateChannelInvite(context.Background(), chID, CreateInviteParams{})
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestGetAnswerVoters(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/polls/456/answers/1" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"789","username":"testuser"}]`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	users, err := c.GetAnswerVoters(context.Background(), chID, msgID, 1, GetAnswerVotersParams{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(users) != 1 || users[0].ID.String() != "789" {
		t.Errorf("Unexpected users: %+v", users)
	}
}

func TestGetAnswerVoters_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	_, err := c.GetAnswerVoters(context.Background(), chID, msgID, 1, GetAnswerVotersParams{})
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}

func TestEndPoll(t *testing.T) {
	ts, c := setupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/channels/123/polls/456/expire" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","content":"ended"}`))
	})
	defer ts.Close()

	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	msg, err := c.EndPoll(context.Background(), chID, msgID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if msg.ID.String() != "456" {
		t.Errorf("Expected msg id 456, got %s", msg.ID.String())
	}
}

func TestEndPoll_Error(t *testing.T) {
	c := setupErrorClient()
	chID, _ := snowflake.Parse("123")
	msgID, _ := snowflake.Parse("456")
	_, err := c.EndPoll(context.Background(), chID, msgID)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
}
