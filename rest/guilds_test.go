package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/discord-go/discord.go/snowflake"
)

func TestCreateGuild(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"Test Guild"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	params := CreateGuildParams{Name: "Test Guild"}
	guild, err := c.CreateGuild(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if guild.ID.String() != "123" || guild.Name != "Test Guild" {
		t.Errorf("Unexpected guild: %+v", guild)
	}
}

func TestModifyGuild(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"123","name":"New Name"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	name := "New Name"
	params := ModifyGuildParams{Name: &name}
	guild, err := c.ModifyGuild(context.Background(), id, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if guild.Name != "New Name" {
		t.Errorf("Unexpected guild name: %s", guild.Name)
	}
}

func TestDeleteGuild(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123" {
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

	id, _ := snowflake.Parse("123")
	err := c.DeleteGuild(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGetGuildChannels(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/channels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"456","name":"general"}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	chs, err := c.GetGuildChannels(context.Background(), id)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(chs) != 1 || chs[0].ID.String() != "456" {
		t.Errorf("Unexpected channels: %+v", chs)
	}
}

func TestCreateGuildChannel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/channels" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"456","name":"new-channel"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	params := CreateGuildChannelParams{Name: "new-channel"}
	ch, err := c.CreateGuildChannel(context.Background(), id, params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ch.Name == nil || *ch.Name != "new-channel" {
		t.Errorf("Unexpected channel: %+v", ch)
	}
}

func TestGetGuildMember(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/members/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user":{"id":"456","username":"testuser"}}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")
	uID, _ := snowflake.Parse("456")
	m, err := c.GetGuildMember(context.Background(), gID, uID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if m.User.ID.String() != "456" {
		t.Errorf("Unexpected member: %+v", m)
	}
}

func TestListGuildMembers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/members" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("Unexpected limit: %s", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"user":{"id":"456"}}]`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	id, _ := snowflake.Parse("123")
	members, err := c.ListGuildMembers(context.Background(), id, ListMembersParams{Limit: 10})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(members) != 1 || members[0].User.ID.String() != "456" {
		t.Errorf("Unexpected members: %+v", members)
	}
}

func TestModifyGuildMember(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/members/456" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"nick":"new-nick"}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")
	uID, _ := snowflake.Parse("456")
	nick := "new-nick"
	m, err := c.ModifyGuildMember(context.Background(), gID, uID, ModifyMemberParams{Nick: &nick})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if m.Nick == nil || *m.Nick != "new-nick" {
		t.Errorf("Unexpected member: %+v", m)
	}
}

func TestRemoveGuildMember(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/members/456" {
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

	gID, _ := snowflake.Parse("123")
	uID, _ := snowflake.Parse("456")
	err := c.RemoveGuildMember(context.Background(), gID, uID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGuildBans(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/guilds/123/bans" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"reason":"spam","user":{"id":"456"}}]`))
		} else if r.URL.Path == "/guilds/123/bans/456" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"reason":"spam","user":{"id":"456"}}`))
		} else if r.URL.Path == "/guilds/123/bans/456" && r.Method == "PUT" {
			w.WriteHeader(http.StatusNoContent)
		} else if r.URL.Path == "/guilds/123/bans/456" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")
	uID, _ := snowflake.Parse("456")

	bans, err := c.GetGuildBans(context.Background(), gID)
	if err != nil || len(bans) != 1 {
		t.Fatalf("Unexpected bans: %v, err: %v", bans, err)
	}

	ban, err := c.GetGuildBan(context.Background(), gID, uID)
	if err != nil || ban.Reason != "spam" {
		t.Fatalf("Unexpected ban: %v, err: %v", ban, err)
	}

	err = c.CreateGuildBan(context.Background(), gID, uID, CreateBanParams{DeleteMessageSeconds: 604800})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	err = c.RemoveGuildBan(context.Background(), gID, uID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGuildRoles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/guilds/123/roles" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":"456","name":"admin"}]`))
		} else if r.URL.Path == "/guilds/123/roles" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"789","name":"mod"}`))
		} else if r.URL.Path == "/guilds/123/roles/789" && r.Method == "PATCH" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":"789","name":"supermod"}`))
		} else if r.URL.Path == "/guilds/123/roles/789" && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")
	rID, _ := snowflake.Parse("789")

	roles, err := c.GetGuildRoles(context.Background(), gID)
	if err != nil || len(roles) != 1 {
		t.Fatalf("Unexpected roles: %v, err: %v", roles, err)
	}

	r, err := c.CreateGuildRole(context.Background(), gID, CreateRoleParams{Name: "mod"})
	if err != nil || r.Name != "mod" {
		t.Fatalf("Unexpected role: %v, err: %v", r, err)
	}

	name := "supermod"
	r2, err := c.ModifyGuildRole(context.Background(), gID, rID, ModifyRoleParams{Name: &name})
	if err != nil || r2.Name != "supermod" {
		t.Fatalf("Unexpected role: %v, err: %v", r2, err)
	}

	err = c.DeleteGuildRole(context.Background(), gID, rID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestGuildPrune(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/guilds/123/prune" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"pruned": 5}`))
		} else if r.URL.Path == "/guilds/123/prune" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"pruned": 10}`))
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")
	count, err := c.GetGuildPruneCount(context.Background(), gID, PruneParams{Days: 7})
	if err != nil || count != 5 {
		t.Fatalf("Unexpected prune count: %d, err: %v", count, err)
	}

	count2, err := c.BeginGuildPrune(context.Background(), gID, PruneParams{Days: 7})
	if err != nil || count2 != 10 {
		t.Fatalf("Unexpected prune begin count: %d, err: %v", count2, err)
	}
}

func TestGuildOtherEndpoints(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/guilds/123/invites" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"code":"abc"}]`))
		} else if r.URL.Path == "/guilds/123/integrations" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":"456"}]`))
		} else if r.URL.Path == "/guilds/123/widget.json" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"enabled":true}`))
		} else {
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")

	invs, err := c.GetGuildInvites(context.Background(), gID)
	if err != nil || len(invs) != 1 {
		t.Fatalf("Unexpected invites: %v, err: %v", invs, err)
	}

	ints, err := c.GetGuildIntegrations(context.Background(), gID)
	if err != nil || len(ints) != 1 {
		t.Fatalf("Unexpected integrations: %v, err: %v", ints, err)
	}

	wid, err := c.GetGuildWidget(context.Background(), gID)
	if err != nil {
		t.Fatalf("Unexpected widget: %v, err: %v", wid, err)
	}
}

func TestGetAuditLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/guilds/123/audit-logs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"audit_log_entries":[{"id":"456"}]}`))
	}))
	defer ts.Close()

	c := New("token", nil, &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return http.DefaultClient.Do(req)
		},
	})
	c.BaseURL = ts.URL

	gID, _ := snowflake.Parse("123")
	al, err := c.GetAuditLog(context.Background(), gID, AuditLogParams{Limit: 10})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(al.AuditLogEntries) != 1 || al.AuditLogEntries[0].ID.String() != "456" {
		t.Errorf("Unexpected audit log: %+v", al)
	}
}
