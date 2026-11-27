package rest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/permissions"
)

func TestComputedPermissionEndpoints(t *testing.T) {
	const highBit permissions.Permission = 1 << 60
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bot token" {
			t.Errorf("invalid method or auth for %s", r.URL.Path)
		}
		switch r.URL.Path {
		case "/guilds/42/permissions/member/7", "/guilds/42/permissions/role/7", "/channels/42/permissions/member/7", "/channels/42/permissions/role/7":
			io.WriteString(w, `{"permissions":"1152921504606846976"}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	c := New("token", nil, server.Client())
	c.BaseURL = server.URL
	ctx := context.Background()
	for name, call := range map[string]func() (permissions.Permission, error){
		"guild member":   func() (permissions.Permission, error) { return c.GetGuildMemberPermissions(ctx, 42, 7) },
		"guild role":     func() (permissions.Permission, error) { return c.GetGuildRolePermissions(ctx, 42, 7) },
		"channel member": func() (permissions.Permission, error) { return c.GetChannelMemberPermissions(ctx, 42, 7) },
		"channel role":   func() (permissions.Permission, error) { return c.GetChannelRolePermissions(ctx, 42, 7) },
	} {
		t.Run(name, func(t *testing.T) {
			got, err := call()
			if err != nil || got != highBit {
				t.Fatalf("permissions = %v, error = %v", got, err)
			}
		})
	}
}
