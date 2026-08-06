package snowflake

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ID
		expectErr bool
	}{
		{"valid zero", "0", ID(0), false},
		{"valid standard", "175928847299117063", ID(175928847299117063), false},
		{"valid max uint64", "18446744073709551615", ID(18446744073709551615), false},
		{"invalid alpha", "invalid", ID(0), true},
		{"invalid empty", "", ID(0), true},
		{"invalid negative", "-1", ID(0), true},
		{"invalid overflow", "18446744073709551616", ID(0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("Parse() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
