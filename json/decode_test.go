package json

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		data    []byte
		want    any
		wantErr bool
	}{
		{
			name:    "valid json",
			data:    []byte(`{"name":"Alice","age":30}`),
			want:    &person{Name: "Alice", Age: 30},
			wantErr: false,
		},
		{
			name:    "invalid json",
			data:    []byte(`{"name":"Alice","age":30`),
			want:    &person{},
			wantErr: true,
		},
		{
			name:    "null",
			data:    []byte(`null`),
			want:    &person{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p person
			err := Unmarshal(tt.data, &p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(&p, tt.want) {
				t.Errorf("Unmarshal() = %v, want %v", &p, tt.want)
			}
		})
	}
}
