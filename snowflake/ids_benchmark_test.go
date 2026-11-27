package snowflake

import (
	"encoding/json"
	"testing"
)

// BenchmarkIDsDecode measures the same 50-ID payload used in local library
// comparisons.
func BenchmarkIDsDecode(b *testing.B) {
	ids := make(IDs, 50)
	for i := range ids {
		ids[i] = ID(123456789012345678 + uint64(i))
	}
	encoded, err := json.Marshal(ids)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded IDs
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			b.Fatal(err)
		}
	}
}
