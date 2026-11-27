package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/discord-go/discord.go/guilds"
	"github.com/discord-go/discord.go/snowflake"
)

func TestGuildWidgetEndpoints(t *testing.T) {
	channel := snowflake.ID(77)
	channelPtr := &channel
	var cleared *snowflake.ID
	disabled := false
	tests := []struct {
		name   string
		params ModifyGuildWidgetParams
		want   string
	}{
		{"omitted", ModifyGuildWidgetParams{}, `{}`},
		{"set", ModifyGuildWidgetParams{Enabled: &disabled, ChannelID: &channelPtr}, `{"enabled":false,"channel_id":"77"}`},
		{"clear", ModifyGuildWidgetParams{ChannelID: &cleared}, `{"channel_id":null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/guilds/42/widget" || r.Header.Get("Authorization") != "Bot token" {
					t.Errorf("unexpected path or auth: %s", r.URL.Path)
				}
				if r.Method == http.MethodPatch {
					body, _ := io.ReadAll(r.Body)
					if string(body) != tt.want {
						t.Errorf("body = %s, want %s", body, tt.want)
					}
					if r.Header.Get("X-Audit-Log-Reason") != "test+reason" {
						t.Error("audit reason missing")
					}
				} else if r.Method != http.MethodGet {
					t.Errorf("unexpected method %s", r.Method)
				}
				io.WriteString(w, `{"enabled":true,"channel_id":"77"}`)
			}))
			defer server.Close()
			c := New("token", nil, server.Client())
			c.BaseURL = server.URL
			got, err := c.GetGuildWidgetSettings(context.Background(), 42)
			if err != nil || !got.Enabled || got.ChannelID == nil || *got.ChannelID != channel {
				t.Fatalf("get: %+v, %v", got, err)
			}
			got, err = c.ModifyGuildWidget(WithReason(context.Background(), "test reason"), 42, tt.params)
			if err != nil || !got.Enabled || got.ChannelID == nil || *got.ChannelID != channel {
				t.Fatalf("modify: %+v, %v", got, err)
			}
		})
	}
}

func TestGuildWidgetImageIsUnauthenticatedAndBinary(t *testing.T) {
	want := []byte{0x89, 'P', 'N', 'G', 0, 0xff}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/guilds/42/widget.png" || r.URL.Query().Get("style") != "banner 1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.String())
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("public widget image sent a token")
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(want)
	}))
	defer server.Close()
	c := New("token", nil, server.Client())
	c.BaseURL = server.URL
	got, err := c.GetGuildWidgetImage(context.Background(), 42, "banner 1")
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestGuildIncidentActionsAndRegions(t *testing.T) {
	until := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	untilPtr := &until
	var cleared *time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bot token" {
			t.Error("missing authorization")
		}
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/guilds/42/incident-actions":
			body, _ := io.ReadAll(r.Body)
			if string(body) != `{"invites_disabled_until":"2026-11-27T12:00:00Z","dms_disabled_until":null}` {
				t.Errorf("body: %s", body)
			}
			io.WriteString(w, `{"invites_disabled_until":"2026-11-27T12:00:00Z","dms_disabled_until":null,"raid_detected_at":"2026-11-27T11:00:00Z"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/guilds/42/regions":
			io.WriteString(w, `[{"id":"vip","name":"VIP","custom":true}]`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	c := New("token", nil, server.Client())
	c.BaseURL = server.URL
	incident, err := c.ModifyGuildIncidentActions(context.Background(), 42, ModifyGuildIncidentActionsParams{InvitesDisabledUntil: &untilPtr, DMsDisabledUntil: &cleared})
	if err != nil || incident.InvitesDisabledUntil == nil || !incident.InvitesDisabledUntil.Equal(until) || incident.DMsDisabledUntil != nil || incident.RaidDetectedAt == nil {
		t.Fatalf("incident: %+v, %v", incident, err)
	}
	regions, err := c.ListGuildVoiceRegions(context.Background(), 42)
	if err != nil || len(regions) != 1 || regions[0].ID != "vip" || !regions[0].Custom {
		t.Fatalf("regions: %+v, %v", regions, err)
	}
	var guild guilds.Guild
	if err := json.Unmarshal([]byte(`{"id":"42","owner_id":"1","incidents_data":{"invites_disabled_until":null,"dms_disabled_until":null,"raid_detected_at":"2026-11-27T11:00:00Z"}}`), &guild); err != nil || guild.IncidentsData == nil || guild.IncidentsData.RaidDetectedAt == nil {
		t.Fatalf("guild: %+v, %v", guild.IncidentsData, err)
	}
}
