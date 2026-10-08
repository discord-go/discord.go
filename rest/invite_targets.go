package rest

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/discord-go/discord.go/snowflake"
)

// Invite target-user job status codes returned by GetInviteTargetUsersJobStatus.
const (
	InviteTargetUsersJobStatusUnspecified = 0
	InviteTargetUsersJobStatusProcessing  = 1
	InviteTargetUsersJobStatusCompleted   = 2
	InviteTargetUsersJobStatusFailed      = 3
)

// InviteTargetUsersJobStatus reports the state of the asynchronous job that
// processes a target-users CSV uploaded at invite creation or through
// UpdateInviteTargetUsers.
type InviteTargetUsersJobStatus struct {
	Status         int        `json:"status"`
	TotalUsers     int        `json:"total_users"`
	ProcessedUsers int        `json:"processed_users"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	ErrorMessage   string     `json:"error_message"`
}

// BulkInviteTargetUsersParams is the JSON body of the bulk add/delete
// target-user endpoints. Discord accepts at most 1000 IDs per request.
type BulkInviteTargetUsersParams struct {
	UserIDs snowflake.IDs `json:"user_ids"`
}

// GetInviteTargetUsers returns the users allowed to see and accept an
// invite. Discord answers with a CSV file whose header is user_id; the rows
// are parsed into snowflakes here. Requires the caller to be the inviter or
// to hold MANAGE_GUILD or VIEW_AUDIT_LOG.
func (c *Client) GetInviteTargetUsers(ctx context.Context, code string) ([]snowflake.ID, error) {
	body, err := c.requestRaw(ctx, "GET", "/invites/"+code+"/target-users", nil)
	if err != nil {
		return nil, err
	}
	return parseTargetUsersCSV(body)
}

// AddInviteTargetUser adds a single user to an invite's target users and
// returns 204 on success. Requires the caller to be the inviter or to hold
// MANAGE_GUILD.
func (c *Client) AddInviteTargetUser(ctx context.Context, code string, userID snowflake.ID) error {
	return c.Request(ctx, "PUT", "/invites/"+code+"/target-users/"+userID.String(), nil, nil)
}

// RemoveInviteTargetUser removes a single user from an invite's target users
// and returns 204 on success. Requires the caller to be the inviter or to
// hold MANAGE_GUILD.
func (c *Client) RemoveInviteTargetUser(ctx context.Context, code string, userID snowflake.ID) error {
	return c.Request(ctx, "DELETE", "/invites/"+code+"/target-users/"+userID.String(), nil, nil)
}

// BulkAddInviteTargetUsers adds up to 1000 users to an invite's target users
// in place and returns 204 on success. Requires the caller to be the inviter
// or to hold MANAGE_GUILD.
func (c *Client) BulkAddInviteTargetUsers(ctx context.Context, code string, params BulkInviteTargetUsersParams) error {
	return c.Request(ctx, "POST", "/invites/"+code+"/target-users/bulk-add", params, nil)
}

// BulkDeleteInviteTargetUsers removes up to 1000 users from an invite's
// target users in place and returns 204 on success. Requires the caller to
// be the inviter or to hold MANAGE_GUILD.
func (c *Client) BulkDeleteInviteTargetUsers(ctx context.Context, code string, params BulkInviteTargetUsersParams) error {
	return c.Request(ctx, "POST", "/invites/"+code+"/target-users/bulk-delete", params, nil)
}

// UpdateInviteTargetUsers replaces an invite's entire target user list. The
// IDs are uploaded as the target_users_file CSV form field; Discord
// processes the file asynchronously, so poll GetInviteTargetUsersJobStatus
// for completion. Unlike the in-place add/remove endpoints this is how the
// list is replaced wholesale. Requires the caller to be the inviter or to
// hold MANAGE_GUILD.
func (c *Client) UpdateInviteTargetUsers(ctx context.Context, code string, userIDs []snowflake.ID) error {
	csvFile, err := buildTargetUsersCSV(userIDs)
	if err != nil {
		return err
	}
	fields := map[string]string{}
	files := []File{{Name: "target_users.csv", ContentType: "text/csv", Reader: bytes.NewReader(csvFile)}}
	return c.RequestMultipartFormNamedFile(ctx, "PUT", "/invites/"+code+"/target-users", fields, "target_users_file", files, nil)
}

// GetInviteTargetUsersJobStatus reports the status of a target-users CSV
// processing job. Requires the caller to be the inviter or to hold
// MANAGE_GUILD or VIEW_AUDIT_LOG.
func (c *Client) GetInviteTargetUsersJobStatus(ctx context.Context, code string) (*InviteTargetUsersJobStatus, error) {
	var status InviteTargetUsersJobStatus
	if err := c.Request(ctx, "GET", "/invites/"+code+"/target-users/job-status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// buildTargetUsersCSV renders the documented CSV layout: a user_id header
// followed by one snowflake per line.
func buildTargetUsersCSV(userIDs []snowflake.ID) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{"user_id"}); err != nil {
		return nil, err
	}
	records := make([][]string, 0, len(userIDs))
	for _, id := range userIDs {
		records = append(records, []string{id.String()})
	}
	writer.WriteAll(records)
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// parseTargetUsersCSV decodes the user_id CSV Discord returns from
// GetInviteTargetUsers.
func parseTargetUsersCSV(body []byte) ([]snowflake.ID, error) {
	reader := csv.NewReader(bytes.NewReader(body))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invite target users: parse csv: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	header := strings.TrimSpace(strings.Join(records[0], ","))
	if !strings.EqualFold(header, "user_id") {
		return nil, fmt.Errorf("invite target users: unexpected csv header %q", header)
	}
	ids := make([]snowflake.ID, 0, len(records)-1)
	for i, record := range records[1:] {
		if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
			continue
		}
		id, err := snowflake.Parse(strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("invite target users: row %d: %w", i+2, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
