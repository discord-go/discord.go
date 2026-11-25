package redislimiter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/discord-go/discord.go/ratelimit"
	"github.com/discord-go/discord.go/rest"
)

func requestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	return ctx
}

// These integration tests use an isolated local Redis when the executables are
// installed. They never connect to or change a user's existing Redis server.
func testRedis(t *testing.T) EvalFunc {
	t.Helper()
	server, err := exec.LookPath("redis-server")
	if err != nil {
		t.Skip("redis-server is not installed")
	}
	cli, err := exec.LookPath("redis-cli")
	if err != nil {
		t.Skip("redis-cli is not installed")
	}
	dir := t.TempDir()
	socket := filepath.Join(dir, "redis.sock")
	cmd := exec.Command(server, "--port", "0", "--unixsocket", socket, "--unixsocketperm", "700", "--dir", dir, "--save", "", "--appendonly", "no")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.DialTimeout("unix", socket, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("test Redis did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return func(ctx context.Context, script string, keys []string, args ...any) (any, error) {
		argv := []string{"-s", socket, "--json", "EVAL", script, strconv.Itoa(len(keys))}
		argv = append(argv, keys...)
		for _, arg := range args {
			argv = append(argv, fmt.Sprint(arg))
		}
		out, err := exec.CommandContext(ctx, cli, argv...).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("redis-cli: %w: %s", err, out)
		}
		var value any
		decoder := json.NewDecoder(bytes.NewReader(out))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		if array, ok := value.([]any); ok {
			for i, v := range array {
				if n, ok := v.(json.Number); ok {
					array[i], _ = n.Int64()
				}
			}
		}
		return value, nil
	}
}

func TestDistributedReservationsAndCooldown(t *testing.T) {
	eval := testRedis(t)
	first, err := New(eval, Options{Namespace: "bot-one"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(eval, Options{Namespace: "bot-one"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := New(eval, Options{Namespace: "bot-two"})
	if err != nil {
		t.Fatal(err)
	}
	path := "/channels/123/messages"
	release, err := first.AcquireRoute(requestContext(t), "GET", path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	if _, err := second.AcquireRoute(ctx, "GET", path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second worker: %v", err)
	}
	cancel()
	otherRelease, err := other.AcquireRoute(requestContext(t), "GET", path)
	if err != nil {
		t.Fatalf("other namespace blocked: %v", err)
	}
	otherRelease()
	if err := first.UpdateRouteContext(context.Background(), "GET", path, ratelimit.Info{Bucket: "bucket", Remaining: 2, ResetAfter: time.Second}); err != nil {
		t.Fatal(err)
	}
	release()
	for _, limiter := range []*Limiter{first, second} {
		done, err := limiter.AcquireRoute(requestContext(t), "GET", path)
		if err != nil {
			t.Fatal(err)
		}
		done()
	}
	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	if _, err := second.AcquireRoute(ctx, "GET", path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("bucket budget: %v", err)
	}
	if err := first.UpdateRouteContext(context.Background(), "GET", path, ratelimit.Info{Global: true, ResetAfter: time.Second}); err != nil {
		t.Fatal(err)
	}
	globalCtx, globalCancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer globalCancel()
	if _, err := second.AcquireRoute(globalCtx, "GET", "/guilds/9"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("global cooldown: %v", err)
	}
}

func TestMajorResourcesAndSecretKeys(t *testing.T) {
	eval := testRedis(t)
	l, err := New(eval, Options{Namespace: "private-bot-token"})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/webhooks/7/secret-one", "/webhooks/7/secret-two"} {
		if err := l.UpdateRouteContext(context.Background(), "POST", path, ratelimit.Info{Bucket: "same", Remaining: 1, ResetAfter: time.Second}); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"/webhooks/7/secret-one", "/webhooks/7/secret-two"} {
		done, err := l.AcquireRoute(requestContext(t), "POST", path)
		if err != nil {
			t.Fatal(err)
		}
		done()
	}
	for _, key := range l.keys {
		if strings.Contains(key, "private-bot-token") || strings.Contains(key, "secret-") {
			t.Fatal("secret appeared in Redis key")
		}
	}
	for _, index := range []int{2, 3, 4} {
		contents, err := eval(context.Background(), fmt.Sprintf("return redis.call('HGETALL', KEYS[%d])", index), l.keys)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(fmt.Sprint(contents), "private-bot-token") || strings.Contains(fmt.Sprint(contents), "secret-") {
			t.Fatal("secret appeared in Redis hash")
		}
	}
}

func TestBackendErrorsFailClosed(t *testing.T) {
	backendErr := errors.New("backend unavailable")
	l, err := New(func(context.Context, string, []string, ...any) (any, error) { return nil, backendErr }, Options{Namespace: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.AcquireRoute(requestContext(t), "GET", "/test"); !errors.Is(err, backendErr) {
		t.Fatalf("got %v", err)
	}
	if err := l.UpdateRouteContext(context.Background(), "GET", "/test", ratelimit.Info{}); !errors.Is(err, backendErr) {
		t.Fatalf("got %v", err)
	}
	if !errors.Is(l.LastError(), backendErr) {
		t.Fatalf("last error = %v", l.LastError())
	}
}

func TestRequestDeadlineMustFitProbeLease(t *testing.T) {
	called := false
	l, err := New(func(context.Context, string, []string, ...any) (any, error) {
		called = true
		return nil, nil
	}, Options{Namespace: "deadline", ProbeLease: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.AcquireRoute(context.Background(), "GET", "/test"); !errors.Is(err, ErrRequestDeadline) {
		t.Fatalf("missing deadline: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := l.AcquireRoute(ctx, "GET", "/test"); !errors.Is(err, ErrRequestDeadline) {
		t.Fatalf("deadline beyond lease: %v", err)
	}
	if called {
		t.Fatal("backend was called without a valid deadline")
	}
}

func TestWorkersRejectMismatchedGlobalLimits(t *testing.T) {
	eval := testRedis(t)
	a, _ := New(eval, Options{Namespace: "settings", GlobalLimit: 50})
	b, _ := New(eval, Options{Namespace: "settings", GlobalLimit: 500})
	release, err := a.AcquireRoute(requestContext(t), "GET", "/test")
	if err != nil {
		t.Fatal(err)
	}
	release()
	if _, err := b.AcquireRoute(requestContext(t), "GET", "/other"); err == nil {
		t.Fatal("worker accepted a different shared global limit")
	}
}

func TestConcurrentWorkersReserveOnlyOneRemainingSlot(t *testing.T) {
	eval := testRedis(t)
	a, _ := New(eval, Options{Namespace: "concurrent"})
	b, _ := New(eval, Options{Namespace: "concurrent"})
	path := "/channels/1/messages"
	if err := a.UpdateRouteContext(context.Background(), "GET", path, ratelimit.Info{Bucket: "shared", Remaining: 1, ResetAfter: time.Second}); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			limiter := a
			if i%2 == 1 {
				limiter = b
			}
			release, err := limiter.AcquireRoute(ctx, "GET", path)
			if err == nil {
				successes.Add(1)
				release()
				return
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("acquire: %v", err)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if got := successes.Load(); got != 1 {
		t.Fatalf("admitted %d workers; want one", got)
	}
}

func TestDistributedGlobalTokenBudget(t *testing.T) {
	eval := testRedis(t)
	a, _ := New(eval, Options{Namespace: "global-budget"})
	b, _ := New(eval, Options{Namespace: "global-budget"})
	_, err := eval(context.Background(), `local t=redis.call('TIME'); local n=t[1]*1000+math.floor(t[2]/1000); for i=1,50 do redis.call('ZADD',KEYS[6],n-750,'seed'..i) end; redis.call('PEXPIRE',KEYS[6],2000); return 1`, a.keys)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, err := b.AcquireRoute(ctx, "GET", "/different"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shared global token budget: %v", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	release, err := b.AcquireRoute(ctx2, "GET", "/different")
	if err != nil {
		t.Fatal(err)
	}
	release()
}

func TestExpiredProbeCanRecoverWithoutStaleReleaseClearingNewOwner(t *testing.T) {
	eval := testRedis(t)
	l, err := New(eval, Options{Namespace: "lease-recovery", ProbeLease: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	firstCtx, firstCancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer firstCancel()
	staleRelease, err := l.AcquireRoute(firstCtx, "GET", "/test")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	secondCtx, secondCancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer secondCancel()
	currentRelease, err := l.AcquireRoute(secondCtx, "GET", "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer currentRelease()
	staleRelease()
	blockedCtx, blockedCancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer blockedCancel()
	if _, err := l.AcquireRoute(blockedCtx, "GET", "/test"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("stale release cleared the current reservation: %v", err)
	}
}

func TestRESTClientsShareDiscoveryReservation(t *testing.T) {
	eval := testRedis(t)
	a, _ := New(eval, Options{Namespace: "rest-workers"})
	b, _ := New(eval, Options{Namespace: "rest-workers"})
	started := make(chan struct{})
	finish := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(started)
			<-finish
		}
		w.Header().Set("X-RateLimit-Bucket", "shared")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset-After", "1")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	first, second := rest.New("test", a, nil), rest.New("test", b, nil)
	first.BaseURL, second.BaseURL = server.URL, server.URL
	done := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() { done <- first.Request(ctx, "GET", "/channels/1/messages", nil, nil) }()
	select {
	case <-started:
	case err := <-done:
		t.Fatalf("first request did not reach server: %v", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel2()
	err := second.Request(ctx2, "GET", "/channels/1/messages", nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("second request: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("sent %d concurrent discovery requests", got)
	}
	close(finish)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
