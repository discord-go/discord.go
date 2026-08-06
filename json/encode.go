package json

import "encoding/json"

// Marshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
