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

func TestMustParse(t *testing.T) {
	id := MustParse("123456789")
	if id != 123456789 {
		t.Errorf("expected 123456789, got %d", id)
	}
}

func TestMustParsePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustParse to panic on invalid input")
		}
	}()
	MustParse("not-a-number")
}

func TestIsZero(t *testing.T) {
	var zero ID
	if !zero.IsZero() {
		t.Error("expected zero ID to be zero")
	}
	if ID(123).IsZero() {
		t.Error("expected non-zero ID to not be zero")
	}
}
