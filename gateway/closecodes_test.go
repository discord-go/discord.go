package gateway

import (
	"testing"
)

func TestCloseCode(t *testing.T) {
	if CloseCodeUnknownError != 4000 {
		t.Error("CloseCodeUnknownError should be 4000")
	}
}
