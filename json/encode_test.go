package json

import (
	"bytes"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    []byte
		wantErr bool
	}{
		{
			name: "valid struct",
			input: struct {
				Name string `json:"name"`
			}{Name: "test"},
			want:    []byte(`{"name":"test"}`),
			wantErr: false,
		},
		{
			name:    "nil slice",
			input:   []string(nil),
			want:    []byte(`null`),
			wantErr: false,
		},
		{
			name:    "empty slice",
			input:   []string{},
			want:    []byte(`[]`),
			wantErr: false,
		},
		{
			name:    "invalid chan",
			input:   make(chan int),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}
