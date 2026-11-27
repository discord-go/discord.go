package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func FuzzMajorResourceCompatibility(f *testing.F) {
	for _, seed := range []string{"", "/channels/1/messages/2", "GET /guilds/2/roles/3", "/webhooks/1/token/messages/2", "/webhooks/1/", "//guilds/1", "/other/1"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, path string) {
		original := path
		if i := strings.IndexByte(path, ' '); i >= 0 {
			path = path[i+1:]
		}
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		want := ""
		if len(parts) >= 2 {
			switch parts[0] {
			case "channels", "guilds":
				want = parts[0] + "/" + parts[1]
			case "webhooks":
				want = parts[0] + "/" + parts[1]
				if len(parts) > 2 {
					want += "/" + parts[2]
				}
			}
		}
		if got := majorResource(original); got != want {
			t.Fatalf("majorResource(%q) = %q, want %q", original, got, want)
		}
	})
}

func TestCollidingRouteLocksKeepDiscoveryReservationsIndependent(t *testing.T) {
	l := NewLimiter(nil).(*limiter)
	seen := make(map[*sync.Mutex]string)
	var first, second string
	for i := 0; i <= len(l.bucketLocks); i++ {
		path := fmt.Sprintf("/channels/%d/messages", i)
		lock := l.getBucketLock("GET " + path)
		if other, ok := seen[lock]; ok {
			first, second = other, path
			break
		}
		seen[lock] = path
	}
	if first == "" {
		t.Fatal("failed to find two routes using the same lock")
	}

	firstRelease, err := l.AcquireRoute(context.Background(), "GET", first)
	if err != nil {
		t.Fatal(err)
	}
	defer firstRelease()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	secondRelease, err := l.AcquireRoute(ctx, "GET", second)
	if err != nil {
		t.Fatalf("unrelated route blocked by an in-flight discovery: %v", err)
	}
	secondRelease()

	blocked, stop := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer stop()
	if release, err := l.AcquireRoute(blocked, "GET", first); !errors.Is(err, context.DeadlineExceeded) {
		if release != nil {
			release()
		}
		t.Fatalf("first route lost its discovery reservation: %v", err)
	}
}

func TestRouteBucketsRespectMajorResourcesAndMethods(t *testing.T) {
	l := NewLimiter(nil).(RouteLimiter)
	l.UpdateRoute("GET", "/channels/1/messages", Info{Bucket: "same", ResetAfter: time.Second})
	l.UpdateRoute("GET", "/channels/2/messages", Info{Bucket: "same", Remaining: 1, ResetAfter: time.Second})
	for _, test := range []struct{ method, path string }{
		{"GET", "/channels/2/messages"},
		{"POST", "/channels/1/messages"},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		err := l.WaitRoute(ctx, test.method, test.path)
		cancel()
		if err != nil {
			t.Fatalf("%s %s unexpectedly blocked: %v", test.method, test.path, err)
		}
	}
	// A second route with the same Discord bucket and major resource shares
	// the cooldown after Discord identifies it.
	l.UpdateRoute("GET", "/channels/1/pins", Info{Bucket: "same", Remaining: 1, ResetAfter: time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.WaitRoute(ctx, "GET", "/channels/1/pins"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shared bucket: got %v", err)
	}
}

func TestWaitCancellationWhileAnotherWaiterIsDelayed(t *testing.T) {
	l := NewLimiter(nil)
	l.Update("/route", Info{ResetAfter: time.Second})
	first, stop := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- l.Wait(first, "/route") }()
	t.Cleanup(func() { stop(); <-done })
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.Wait(ctx, "/route"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want deadline exceeded", err)
	}
}

func TestGlobalResetAfter(t *testing.T) {
	l := NewLimiter(nil)
	l.Update("", Info{Global: true, ResetAfter: time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.Wait(ctx, "/unrelated"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want global cooldown", err)
	}
}

func TestHeaderless429UpdatesKnownBucket(t *testing.T) {
	l := NewLimiter(nil).(RouteLimiter)
	l.UpdateRoute("POST", "/webhooks/1/first", Info{Bucket: "same", Remaining: 3, ResetAfter: time.Second})
	l.UpdateRoute("POST", "/webhooks/1/first", Info{ResetAfter: time.Second})
	l.UpdateRoute("GET", "/webhooks/1/first/messages", Info{Bucket: "same", Remaining: 3, ResetAfter: time.Second})
	l.UpdateRoute("POST", "/webhooks/1/second", Info{Bucket: "same", Remaining: 1, ResetAfter: time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.WaitRoute(ctx, "GET", "/webhooks/1/first/messages"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("known bucket ignored cooldown: %v", err)
	}
	if err := l.WaitRoute(context.Background(), "POST", "/webhooks/1/second"); err != nil {
		t.Fatalf("different webhook token unexpectedly blocked: %v", err)
	}
}

func TestOutOfOrderUpdatesAndNewWindow(t *testing.T) {
	l := NewLimiter(nil)
	route := l.(RouteLimiter)
	reader := l.(StateReader)
	reset := time.Now().Add(250 * time.Millisecond)
	route.UpdateRoute("GET", "/channels/1/messages", Info{Bucket: "b", Remaining: 0, Reset: reset})
	route.UpdateRoute("GET", "/channels/1/messages", Info{Bucket: "b", Remaining: 4, ResetAfter: 10 * time.Millisecond})
	state, ok := reader.State("/channels/1/messages")
	if !ok || state.Remaining != 0 || !state.Reset.Equal(reset) {
		t.Fatalf("stale update reopened the bucket: %+v", state)
	}
	time.Sleep(time.Until(state.Reset) + 5*time.Millisecond)
	route.UpdateRoute("GET", "/channels/1/messages", Info{Bucket: "b", Remaining: 3, ResetAfter: time.Second})
	state, ok = reader.State("/channels/1/messages")
	if !ok || state.Remaining != 3 {
		t.Fatalf("new window ignored: %+v", state)
	}
}
