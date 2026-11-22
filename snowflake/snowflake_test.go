package snowflake

import "testing"

func TestID_String(t *testing.T) {
	tests := []struct {
		id       ID
		expected string
	}{
		{ID(0), "0"},
		{ID(175928847299117063), "175928847299117063"},
		{ID(18446744073709551615), "18446744073709551615"},
	}

	for _, tt := range tests {
		if got := tt.id.String(); got != tt.expected {
			t.Errorf("ID.String() = %v, want %v", got, tt.expected)
		}
	}
}

func TestIDMarshalJSONBoundaries(t *testing.T) {
	for _, tt := range []struct {
		id   ID
		want string
	}{
		{0, `"0"`}, {1, `"1"`}, {9007199254740993, `"9007199254740993"`}, {^ID(0), `"18446744073709551615"`},
	} {
		got, err := tt.id.MarshalJSON()
		if err != nil || string(got) != tt.want {
			t.Errorf("ID(%d) = %s, %v; want %s", tt.id, got, err, tt.want)
		}
	}
}
