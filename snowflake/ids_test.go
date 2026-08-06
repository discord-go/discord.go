package snowflake

import (
	"encoding/json"
	"testing"
)

func TestIDsUseDiscordStringEncoding(t *testing.T) {
	encoded, err := json.Marshal(IDs{1, 9007199254740993})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `["1","9007199254740993"]` {
		t.Fatalf("encoded IDs = %s", encoded)
	}
	var decoded IDs
	if err := json.Unmarshal([]byte(`["1",9007199254740993]`), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 2 || decoded[1].String() != "9007199254740993" {
		t.Fatalf("decoded IDs = %#v", decoded)
	}
}
