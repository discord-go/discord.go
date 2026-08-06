package snowflake

import (
	"encoding/json"
	"fmt"
)

// IDs is a Discord API-compatible list of snowflakes. Discord serializes
// snowflake arrays as strings even when individual IDs are represented by ID.
type IDs []ID

func (ids IDs) MarshalJSON() ([]byte, error) {
	values := make([]string, len(ids))
	for index, id := range ids {
		values[index] = id.String()
	}
	return json.Marshal(values)
}

func (ids *IDs) UnmarshalJSON(data []byte) error {
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	result := make(IDs, len(values))
	for index, raw := range values {
		var stringValue string
		if err := json.Unmarshal(raw, &stringValue); err == nil {
			id, parseErr := Parse(stringValue)
			if parseErr != nil {
				return parseErr
			}
			result[index] = id
			continue
		}
		var number uint64
		if err := json.Unmarshal(raw, &number); err != nil {
			return fmt.Errorf("snowflake: invalid ID at index %d: %w", index, err)
		}
		result[index] = ID(number)
	}
	*ids = result
	return nil
}
