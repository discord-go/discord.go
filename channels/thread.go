package channels

import (
	"time"

	"github.com/discord-go/discord.go/snowflake"
	"github.com/discord-go/discord.go/users"
)

type ThreadMetadata struct {
	Archived            bool       `json:"archived"`
	AutoArchiveDuration int        `json:"auto_archive_duration"`
	ArchiveTimestamp    time.Time  `json:"archive_timestamp"`
	Locked              bool       `json:"locked"`
	Invitable           bool       `json:"invitable,omitempty"`
	CreateTimestamp     *time.Time `json:"create_timestamp,omitempty"`
}

type ThreadMember struct {
	ID            *snowflake.ID `json:"id,string,omitempty"`
	UserID        *snowflake.ID `json:"user_id,string,omitempty"`
	JoinTimestamp time.Time     `json:"join_timestamp"`
	Flags         int           `json:"flags"`
	Member        *users.Member `json:"member,omitempty"`
}
