package messages

import (
	"encoding/json"
	"time"

	"github.com/discord-go/discord.go/snowflake"
)

// MessageCall represents a call object in a message.
type MessageCall struct {
	Participants   []snowflake.ID `json:"participants"`
	EndedTimestamp *time.Time     `json:"ended_timestamp,omitempty"`
}

// UnmarshalJSON handles unmarshaling array of snowflake strings.
func (c *MessageCall) UnmarshalJSON(data []byte) error {
	type Alias MessageCall
	aux := &struct {
		Participants []string `json:"participants"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Participants != nil {
		c.Participants = make([]snowflake.ID, len(aux.Participants))
		for i, p := range aux.Participants {
			id, err := snowflake.Parse(p)
			if err != nil {
				return err
			}
			c.Participants[i] = id
		}
	}
	return nil
}
