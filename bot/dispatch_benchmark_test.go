package bot

import (
	"encoding/json"
	"testing"
)

// BenchmarkDispatchEventUnsubscribed measures the generic-subscription lookup
// when a bot receives an event without a generic listener. It excludes JSON
// decoding, registered handler execution, and network traffic.
func BenchmarkDispatchEventUnsubscribed(b *testing.B) {
	bot := &Bot{}
	data := json.RawMessage(`{"id":"123","content":"hello"}`)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bot.dispatchEvent("MESSAGE_CREATE", data)
		}
	})
}

// BenchmarkMessageCreateDispatch includes gateway payload and message decoding
// without a cache, router, or user handlers. It does not measure network I/O.
func BenchmarkMessageCreateDispatch(b *testing.B) {
	bot := New("benchmark")
	data := []byte(`{"op":0,"t":"MESSAGE_CREATE","s":1,"d":{"id":"123456789012345678","channel_id":"223456789012345678","guild_id":"423456789012345678","author":{"id":"323456789012345678","username":"Alice","discriminator":"0"},"content":"hello world","timestamp":"2025-01-02T03:04:05Z","tts":false,"mention_everyone":false,"mentions":[],"mention_roles":[],"attachments":[],"embeds":[],"pinned":false,"type":0}}`)
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bot.handleRawDispatch(data)
	}
}
