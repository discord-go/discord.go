package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGuildTemplateEndpoints(t *testing.T) {
	var wantMethod, wantPath, wantBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != wantMethod || r.URL.EscapedPath() != wantPath {
			t.Errorf("got %s %s; want %s %s", r.Method, r.URL.EscapedPath(), wantMethod, wantPath)
		}
		if r.Header.Get("Authorization") != "Bot token" {
			t.Error("missing authorization")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != wantBody {
			t.Errorf("body = %q, want %q", body, wantBody)
		}
		if r.Method == "GET" && r.URL.Path == "/guilds/42/templates" {
			io.WriteString(w, `[{"code":"abc"}]`)
			return
		}
		io.WriteString(w, `{"code":"abc","name":"example"}`)
	}))
	defer server.Close()
	c := New("token", nil, &testHTTPClient{})
	c.BaseURL = server.URL
	ctx := context.Background()
	tests := []struct {
		method, path, body string
		call               func() error
	}{
		{"GET", "/guilds/templates/a%2Fb", "", func() error { _, err := c.GetGuildTemplate(ctx, "a/b"); return err }},
		{"GET", "/guilds/42/templates", "", func() error {
			result, err := c.GetGuildTemplates(ctx, 42)
			if err == nil && len(result) != 1 {
				t.Error("missing template")
			}
			return err
		}},
		{"POST", "/guilds/42/templates", `{"name":"example"}`, func() error {
			_, err := c.CreateGuildTemplate(ctx, 42, CreateGuildTemplateParams{Name: "example"})
			return err
		}},
		{"PUT", "/guilds/42/templates/abc", "", func() error { _, err := c.SyncGuildTemplate(ctx, 42, "abc"); return err }},
		{"PATCH", "/guilds/42/templates/abc", `{"description":""}`, func() error {
			empty := ""
			_, err := c.ModifyGuildTemplate(ctx, 42, "abc", ModifyGuildTemplateParams{Description: &empty})
			return err
		}},
		{"DELETE", "/guilds/42/templates/abc", "", func() error {
			result, err := c.DeleteGuildTemplate(ctx, 42, "abc")
			if err == nil && result.Code != "abc" {
				t.Error("missing deleted template")
			}
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.method+strings.ReplaceAll(test.path, "/", "_"), func(t *testing.T) {
			wantMethod, wantPath, wantBody = test.method, test.path, test.body
			if err := test.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIntegrationDeletionAndStickerPack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/guilds/42/integrations/9":
			if r.Header.Get("X-Audit-Log-Reason") != "cleanup" {
				t.Error("missing audit reason")
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/sticker-packs/7":
			io.WriteString(w, `{"id":"7","name":"pack","sku_id":"8","stickers":[],"banner_asset_id":"10","cover_sticker_id":"11"}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client := New("token", nil, &testHTTPClient{})
	client.BaseURL = server.URL
	if err := client.DeleteGuildIntegration(WithReason(context.Background(), "cleanup"), 42, 9); err != nil {
		t.Fatal(err)
	}
	pack, err := client.GetStickerPack(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if pack.Name != "pack" || pack.BannerAsset == nil || *pack.BannerAsset != "10" || pack.CoverAsset == nil || *pack.CoverAsset != "11" {
		t.Errorf("incorrect pack: %+v", pack)
	}
}
