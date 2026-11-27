package guilds

import "time"

// IncidentsData describes a guild's current incident actions and detections.
type IncidentsData struct {
	InvitesDisabledUntil *time.Time `json:"invites_disabled_until"`
	DMsDisabledUntil     *time.Time `json:"dms_disabled_until"`
	DMSpamDetectedAt     *time.Time `json:"dm_spam_detected_at,omitempty"`
	RaidDetectedAt       *time.Time `json:"raid_detected_at,omitempty"`
}
