package json

import "encoding/json"

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
