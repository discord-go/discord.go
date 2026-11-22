package snowflake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// IDs is a Discord API-compatible list of snowflakes. Discord serializes
// snowflake arrays as strings even when individual IDs are represented by ID.
type IDs []ID

func (ids IDs) MarshalJSON() ([]byte, error) {
	encoded := make([]byte, 0, 2+len(ids)*23)
	encoded = append(encoded, '[')
	for index, id := range ids {
		if index > 0 {
			encoded = append(encoded, ',')
		}
		encoded = append(encoded, '"')
		encoded = strconv.AppendUint(encoded, uint64(id), 10)
		encoded = append(encoded, '"')
	}
	return append(encoded, ']'), nil
}

func (ids *IDs) UnmarshalJSON(data []byte) error {
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	result := make(IDs, len(values))
	for index, raw := range values {
		var stringValue string
		if len(raw) > 0 && raw[0] == '"' {
			if bytes.IndexByte(raw, '\\') < 0 {
				stringValue = string(raw[1 : len(raw)-1])
			} else if err := json.Unmarshal(raw, &stringValue); err != nil {
				return fmt.Errorf("snowflake: invalid ID at index %d: %w", index, err)
			}
			id, parseErr := Parse(stringValue)
			if parseErr != nil {
				return parseErr
			}
			result[index] = id
			continue
		}
		// Unmarshalling null into a string produces an empty string; keep the
		// existing invalid-ID behavior instead of accepting it as numeric zero.
		if bytes.Equal(raw, []byte("null")) {
			_, err := Parse("")
			return err
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
