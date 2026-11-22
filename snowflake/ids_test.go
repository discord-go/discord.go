package snowflake

import (
	"encoding/json"
	"reflect"
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

func TestIDsMarshalJSONBoundaries(t *testing.T) {
	cases := []struct {
		ids  IDs
		want string
	}{
		{nil, `[]`}, {IDs{}, `[]`}, {IDs{0, ^ID(0)}, `["0","18446744073709551615"]`},
	}
	for _, tt := range cases {
		got, err := json.Marshal(tt.ids)
		if err != nil || string(got) != tt.want {
			t.Errorf("Marshal(%v) = %s, %v; want %s", tt.ids, got, err, tt.want)
		}
	}
}

func TestIDsUnmarshalPreservesFormatsAndErrors(t *testing.T) {
	var ids IDs
	if err := json.Unmarshal([]byte(`["\u0031","2","0003",4,18446744073709551615]`), &ids); err != nil {
		t.Fatal(err)
	}
	want := IDs{1, 2, 3, 4, ^ID(0)}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("got %v, want %v", ids, want)
	}
	for _, input := range []string{`[null]`, `[true]`, `[1.5]`, `["-1"]`, `["+2"]`, `["18446744073709551616"]`, `["1",`} {
		if err := json.Unmarshal([]byte(input), &ids); err == nil {
			t.Errorf("expected error for %s", input)
		}
	}
}

// Verify the optimized parser against the previous decoding behavior.
func TestIDsDecodeCompatibility(t *testing.T) {
	inputs := []string{
		`null`, `[]`, `["1",2]`, `["\u0031","0002"]`, `[0,18446744073709551615]`,
		`[null]`, `[false]`, `[{}]`, `[1.5]`, `[-1]`, `["-1"]`, `["+2"]`, `[""]`,
		`[18446744073709551616]`, `["18446744073709551616"]`, `"1"`, `["1",`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			want, wantErr := decodeIDsReference([]byte(input))
			got := IDs{999}
			gotErr := got.UnmarshalJSON([]byte(input))
			if (wantErr != nil) != (gotErr != nil) {
				t.Fatalf("error = %v; reference = %v", gotErr, wantErr)
			}
			if gotErr == nil && !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v; reference = %v", got, want)
			}
			if gotErr != nil && !reflect.DeepEqual(got, IDs{999}) {
				t.Fatalf("receiver changed after error: %v", got)
			}
		})
	}
}

func decodeIDsReference(data []byte) (IDs, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	result := make(IDs, len(raw))
	for i, item := range raw {
		var text string
		if err := json.Unmarshal(item, &text); err == nil {
			id, err := Parse(text)
			if err != nil {
				return nil, err
			}
			result[i] = id
			continue
		}
		var number uint64
		if err := json.Unmarshal(item, &number); err != nil {
			return nil, err
		}
		result[i] = ID(number)
	}
	return result, nil
}
