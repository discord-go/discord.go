package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/discord-go/discord.go/components"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/discord-go/discord.go/ratelimit"
	"github.com/discord-go/discord.go/snowflake"
)

func TestCreateInteractionResponseWithFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("interaction callback must not use bot authorization")
		}
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatal(err)
		}
		payload := r.FormValue("payload_json")
		if !strings.Contains(payload, `"flags":32768`) || !strings.Contains(payload, `"filename":"export.json"`) {
			t.Fatalf("payload_json = %s", payload)
		}
		file, _, err := r.FormFile("files[0]")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		content, _ := io.ReadAll(file)
		if string(content) != "hello" {
			t.Fatalf("file content = %q", content)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := New("token", ratelimit.NewLimiter(ratelimit.NewMemoryStore()), &testHTTPClient{})
	c.BaseURL = server.URL
	response := interactions.InteractionResponse{
		Type: interactions.InteractionCallbackTypeChannelMessageWithSource,
		Data: &interactions.InteractionCallbackData{
			Flags:      messages.FlagIsComponentsV2,
			Components: []components.Component{components.TextDisplay{Content: "hello"}},
		},
	}
	if err := c.CreateInteractionResponseWithFiles(context.Background(), snowflake.ID(1), "token", response, []File{FileFromBytes("export.json", []byte("hello"))}); err != nil {
		t.Fatal(err)
	}
}

type testHTTPClient struct{}

func (*testHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}
func (*testHTTPClient) Get(string) (*http.Response, error) { return nil, nil }
func (*testHTTPClient) Post(string, string, io.Reader) (*http.Response, error) {
	return nil, nil
}
