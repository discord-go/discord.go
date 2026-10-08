package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/ratelimit"
	"github.com/discord-go/discord.go/snowflake"
)

func newInviteTargetClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := New("token", ratelimit.NewLimiter(ratelimit.NewMemoryStore()), &testHTTPClient{})
	c.BaseURL = ts.URL
	return c
}

func TestGetInviteTargetUsers(t *testing.T) {
	var gotMethod, gotPath string
	c := newInviteTargetClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "text/csv")
		w.Write([]byte("user_id\n123456789012345678\n987654321098765432\n"))
	})

	ids, err := c.GetInviteTargetUsers(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/invites/abc123/target-users" {
		t.Errorf("path = %q", gotPath)
	}
	if len(ids) != 2 {
		t.Fatalf("len(ids) = %d, want 2", len(ids))
	}
	if ids[0] != snowflake.ID(123456789012345678) || ids[1] != snowflake.ID(987654321098765432) {
		t.Errorf("ids = %v", ids)
	}
}

func TestGetInviteTargetUsers_MalformedCSV(t *testing.T) {
	c := newInviteTargetClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("nope\n1\n"))
	})
	if _, err := c.GetInviteTargetUsers(context.Background(), "abc123"); err == nil {
		t.Fatal("expected an error for an unexpected csv header")
	}
}

func TestAddAndRemoveInviteTargetUser(t *testing.T) {
	type call struct{ method, path string }
	var calls []call
	c := newInviteTargetClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, call{r.Method, r.URL.Path})
		w.WriteHeader(http.StatusNoContent)
	})

	const code = "abc123"
	userID := snowflake.ID(123456789012345678)
	if err := c.AddInviteTargetUser(context.Background(), code, userID); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveInviteTargetUser(context.Background(), code, userID); err != nil {
		t.Fatal(err)
	}

	want := []call{
		{http.MethodPut, "/invites/abc123/target-users/123456789012345678"},
		{http.MethodDelete, "/invites/abc123/target-users/123456789012345678"},
	}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v", calls)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("call %d = %v, want %v", i, calls[i], want[i])
		}
	}
}

func TestBulkInviteTargetUsers(t *testing.T) {
	type call struct {
		method, path string
		body         string
	}
	var calls []call
	c := newInviteTargetClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		calls = append(calls, call{r.Method, r.URL.Path, string(raw)})
		w.WriteHeader(http.StatusNoContent)
	})

	params := BulkInviteTargetUsersParams{UserIDs: snowflake.IDs{1, 2}}
	if err := c.BulkAddInviteTargetUsers(context.Background(), "abc123", params); err != nil {
		t.Fatal(err)
	}
	if err := c.BulkDeleteInviteTargetUsers(context.Background(), "abc123", params); err != nil {
		t.Fatal(err)
	}

	if len(calls) != 2 {
		t.Fatalf("calls = %v", calls)
	}
	if calls[0].method != http.MethodPost || calls[0].path != "/invites/abc123/target-users/bulk-add" {
		t.Errorf("call 0 = %+v", calls[0])
	}
	if calls[1].method != http.MethodPost || calls[1].path != "/invites/abc123/target-users/bulk-delete" {
		t.Errorf("call 1 = %+v", calls[1])
	}
	var decoded BulkInviteTargetUsersParams
	if err := json.Unmarshal([]byte(calls[0].body), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.UserIDs) != 2 || decoded.UserIDs[0] != 1 || decoded.UserIDs[1] != 2 {
		t.Errorf("body = %s", calls[0].body)
	}
}

func TestUpdateInviteTargetUsers(t *testing.T) {
	var gotMethod, gotPath, contentType string
	var csvBody []byte
	c := newInviteTargetClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		contentType = r.Header.Get("Content-Type")
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatal(err)
		}
		file, header, err := r.FormFile("target_users_file")
		if err != nil {
			t.Fatalf("target_users_file: %v", err)
		}
		defer file.Close()
		if header.Filename != "target_users.csv" {
			t.Errorf("filename = %q", header.Filename)
		}
		csvBody, _ = io.ReadAll(file)
		w.WriteHeader(http.StatusNoContent)
	})

	userIDs := []snowflake.ID{123456789012345678, 987654321098765432}
	if err := c.UpdateInviteTargetUsers(context.Background(), "abc123", userIDs); err != nil {
		t.Fatal(err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/invites/abc123/target-users" {
		t.Errorf("path = %q", gotPath)
	}
	if contentType == "" {
		t.Error("missing content type")
	}
	want := "user_id\n123456789012345678\n987654321098765432\n"
	if string(csvBody) != want {
		t.Errorf("csv = %q, want %q", csvBody, want)
	}
}

func TestGetInviteTargetUsersJobStatus(t *testing.T) {
	c := newInviteTargetClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q", r.Method)
		}
		if r.URL.Path != "/invites/abc123/target-users/job-status" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(`{"status":1,"total_users":100,"processed_users":40,"created_at":"2026-09-01T12:00:00Z","completed_at":null,"error_message":""}`))
	})

	status, err := c.GetInviteTargetUsersJobStatus(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != InviteTargetUsersJobStatusProcessing {
		t.Errorf("status = %d", status.Status)
	}
	if status.TotalUsers != 100 || status.ProcessedUsers != 40 {
		t.Errorf("counts = %+v", status)
	}
	if status.CompletedAt != nil {
		t.Errorf("completed_at = %v, want nil", status.CompletedAt)
	}
}

func TestTargetUsersCSVRoundTrip(t *testing.T) {
	ids := []snowflake.ID{1, 2, 3}
	raw, err := buildTargetUsersCSV(ids)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseTargetUsersCSV(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 3 || parsed[0] != 1 || parsed[2] != 3 {
		t.Errorf("parsed = %v", parsed)
	}
}

func TestCreateInviteParamsTargetUserIDs(t *testing.T) {
	raw, err := json.Marshal(CreateInviteParams{TargetUserIDs: snowflake.IDs{7, 8}})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	list, ok := decoded["target_user_ids"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("target_user_ids = %#v", decoded["target_user_ids"])
	}
	if list[0] != "7" || list[1] != "8" {
		t.Errorf("target_user_ids = %v", list)
	}
}

func TestAPIErrorKeepsBareObjectBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"target_users_file":["513245678901234567"]}`))
	}))
	defer ts.Close()

	c := New("token", ratelimit.NewLimiter(ratelimit.NewMemoryStore()), &testHTTPClient{})
	c.BaseURL = ts.URL
	err := c.UpdateInviteTargetUsers(context.Background(), "abc123", []snowflake.ID{513245678901234567})
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if apiErr.HTTPStatus != http.StatusBadRequest {
		t.Errorf("http status = %d", apiErr.HTTPStatus)
	}
	if apiErr.Message == "" {
		t.Error("expected the response body to be preserved in the error message")
	}
}
