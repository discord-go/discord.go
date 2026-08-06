package gateway

import (
	"encoding/json"
	"testing"
)

func TestResume_JSON(t *testing.T) {
	r := Resume{
		Token:     "token",
		SessionID: "session",
		Seq:       10,
	}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	var r2 Resume
	err = json.Unmarshal(b, &r2)
	if err != nil {
		t.Fatal(err)
	}

	if r2.Token != r.Token {
		t.Errorf("expected %s, got %s", r.Token, r2.Token)
	}
}
