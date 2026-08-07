package snowflake

import "testing"

func BenchmarkID_String(b *testing.B) {
	id := ID(175928847299117063)
	for i := 0; i < b.N; i++ {
		_ = id.String()
	}
}

func BenchmarkParse(b *testing.B) {
	s := "175928847299117063"
	for i := 0; i < b.N; i++ {
		_, _ = Parse(s)
	}
}
