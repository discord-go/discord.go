package bot

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentEventDispatchInvokesOnceOnlyOnce(t *testing.T) {
	b := New("test")
	var onceCalls, regularCalls atomic.Int32
	b.OnceEvent("CUSTOM", func(*EventContext) { onceCalls.Add(1) })
	b.OnEvent("CUSTOM", func(*EventContext) { regularCalls.Add(1) })
	const dispatchers = 64
	start := make(chan struct{})
	var dispatches sync.WaitGroup
	for i := 0; i < dispatchers; i++ {
		dispatches.Add(1)
		go func() {
			defer dispatches.Done()
			<-start
			b.dispatchEvent("CUSTOM", json.RawMessage(`{"ok":true}`))
		}()
	}
	close(start)
	dispatches.Wait()
	b.handlerWG.Wait()
	if got := onceCalls.Load(); got != 1 {
		t.Fatalf("one-shot handler ran %d times", got)
	}
	if got := regularCalls.Load(); got != dispatchers {
		t.Fatalf("regular handler ran %d times, want %d", got, dispatchers)
	}
}
