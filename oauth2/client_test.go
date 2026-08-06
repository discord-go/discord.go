package oauth2

import (
	"net/url"
	"testing"
)

func TestAuthorizationURL(t *testing.T) {
	client := New(Config{ClientID: "123", RedirectURI: "https://example.test/callback"})
	parsed, err := url.Parse(client.AuthorizationURL([]string{"identify", "guilds"}, "state"))
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if query.Get("client_id") != "123" || query.Get("scope") != "identify guilds" || query.Get("state") != "state" {
		t.Fatalf("authorization query = %v", query)
	}
}
