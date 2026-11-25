// Package redislimiter coordinates Discord REST rate limits through Redis.
package redislimiter

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/discord-go/discord.go/ratelimit"
)

// EvalFunc executes a Redis EVAL command and returns its decoded reply. For
// example, a go-redis client can be adapted by returning client.Eval(...).Result().
type EvalFunc func(ctx context.Context, script string, keys []string, args ...any) (any, error)

// ErrRequestDeadline indicates that a request may outlive its probe lease.
var ErrRequestDeadline = errors.New("redislimiter: request context requires a deadline shorter than ProbeLease")

// Options configures shared coordination. Every process using the same bot
// must use the same namespace and Redis database.
type Options struct {
	Namespace string
	// GlobalLimit is the maximum requests in a rolling second. It defaults to
	// 50; increase it only when Discord has granted the bot a higher limit.
	GlobalLimit int
	// ProbeLease bounds how long an abandoned discovery request blocks a route.
	// HTTP request deadlines must be shorter than this lease.
	ProbeLease time.Duration
	// OperationTimeout bounds each Redis call.
	OperationTimeout time.Duration
	// CacheTTL controls how long idle route and bucket metadata is retained.
	CacheTTL time.Duration
}

// Limiter uses atomic Redis scripts to share global and route reservations.
type Limiter struct {
	eval                     EvalFunc
	keys                     []string
	lease, timeout, cacheTTL time.Duration
	globalLimit              int
	errMu                    sync.Mutex
	lastErr                  error
}

var _ ratelimit.Limiter = (*Limiter)(nil)
var _ ratelimit.RouteLimiter = (*Limiter)(nil)
var _ ratelimit.RouteReserver = (*Limiter)(nil)
var _ ratelimit.ContextRouteUpdater = (*Limiter)(nil)

// New creates a distributed limiter. Namespace must identify one bot's shared
// request budget; do not reuse it for different bot tokens.
func New(eval EvalFunc, options Options) (*Limiter, error) {
	if eval == nil || options.Namespace == "" {
		return nil, fmt.Errorf("redislimiter: evaluator and namespace are required")
	}
	if options.ProbeLease == 0 {
		options.ProbeLease = 5 * time.Minute
	}
	if options.GlobalLimit == 0 {
		options.GlobalLimit = 50
	}
	if options.OperationTimeout == 0 {
		options.OperationTimeout = 3 * time.Second
	}
	if options.CacheTTL == 0 {
		options.CacheTTL = time.Hour
	}
	if options.GlobalLimit < 1 || options.ProbeLease < time.Second || options.OperationTimeout <= 0 || options.CacheTTL < time.Minute {
		return nil, fmt.Errorf("redislimiter: invalid lease, timeout, or cache TTL")
	}
	// The hash tag places all script keys in one Redis Cluster slot. Route and
	// namespace values may contain secrets, so only their hashes are stored.
	prefix := "discordgo:ratelimit:{" + digest(options.Namespace) + "}:"
	return &Limiter{eval: eval, keys: []string{prefix + "global", prefix + "routes", prefix + "buckets", prefix + "probes", prefix + "expiry", prefix + "global-window"},
		lease: options.ProbeLease, timeout: options.OperationTimeout, cacheTTL: options.CacheTTL, globalLimit: options.GlobalLimit}, nil
}

func digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func routeID(method, path string) string { return digest(strings.ToUpper(method) + " " + path) }

func majorResource(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	switch parts[0] {
	case "channels", "guilds":
		return parts[0] + "/" + parts[1]
	case "webhooks":
		if len(parts) > 2 {
			return parts[0] + "/" + parts[1] + "/" + parts[2]
		}
		return parts[0] + "/" + parts[1]
	default:
		return ""
	}
}

func (l *Limiter) remember(err error) {
	l.errMu.Lock()
	l.lastErr = err
	l.errMu.Unlock()
}

// LastError reports the most recent backend or reservation-release error.
func (l *Limiter) LastError() error {
	l.errMu.Lock()
	defer l.errMu.Unlock()
	return l.lastErr
}

// AcquireRoute atomically reserves global capacity and route capacity.
func (l *Limiter) AcquireRoute(ctx context.Context, method, path string) (func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) >= l.lease {
		return nil, ErrRequestDeadline
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil, fmt.Errorf("redislimiter: reservation ID: %w", err)
	}
	owner := hex.EncodeToString(random[:])
	route := routeID(method, path)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		opCtx, cancel := context.WithTimeout(ctx, l.timeout)
		result, err := l.eval(opCtx, acquireScript, l.keys, route, owner, l.lease.Milliseconds(), l.cacheTTL.Milliseconds(), l.globalLimit)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			l.remember(err)
			return nil, fmt.Errorf("redislimiter: acquire: %w", err)
		}
		values, ok := result.([]any)
		if !ok || len(values) != 2 {
			return nil, fmt.Errorf("redislimiter: unexpected acquire reply")
		}
		delay, err := integer(values[0])
		if err != nil {
			return nil, err
		}
		if delay < 0 {
			return nil, fmt.Errorf("redislimiter: invalid negative wait")
		}
		probe, ok := values[1].(string)
		if !ok {
			return nil, fmt.Errorf("redislimiter: unexpected probe reply")
		}
		if delay == 0 {
			if probe == "" {
				return func() {}, nil
			}
			var once sync.Once
			return func() {
				once.Do(func() {
					ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
					defer cancel()
					_, err := l.eval(ctx, releaseScript, l.keys, probe, owner)
					if err != nil {
						l.remember(fmt.Errorf("redislimiter: release: %w", err))
					}
				})
			}, nil
		}
		// Polling is bounded so release of an unknown bucket wakes other
		// workers promptly without requiring Redis pub/sub connections.
		if delay > 100 {
			delay = 100
		}
		timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func integer(v any) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		if err == nil {
			return i, nil
		}
	}
	return 0, fmt.Errorf("redislimiter: unexpected integer reply")
}

// Wait is the legacy interface; lifecycle-aware callers should use AcquireRoute.
func (l *Limiter) Wait(ctx context.Context, bucket string) error {
	release, err := l.AcquireRoute(ctx, "", bucket)
	if err == nil {
		release()
	}
	return err
}

// WaitRoute is the legacy route interface; REST clients use AcquireRoute.
func (l *Limiter) WaitRoute(ctx context.Context, method, path string) error {
	release, err := l.AcquireRoute(ctx, method, path)
	if err == nil {
		release()
	}
	return err
}

// Update records a response through the legacy interface. Use LastError to
// inspect an error because the Limiter interface cannot return one.
func (l *Limiter) Update(bucket string, info ratelimit.Info) { l.UpdateRoute("", bucket, info) }

// UpdateRoute records a response through the legacy route interface.
func (l *Limiter) UpdateRoute(method, path string, info ratelimit.Info) {
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()
	if err := l.UpdateRouteContext(ctx, method, path, info); err != nil {
		l.remember(err)
	}
}

// UpdateRouteContext atomically records Discord's response for all workers.
func (l *Limiter) UpdateRouteContext(ctx context.Context, method, path string, info ratelimit.Info) error {
	delay := info.ResetAfter
	if delay <= 0 && !info.Reset.IsZero() {
		delay = time.Until(info.Reset)
	}
	if delay < 0 {
		delay = 0
	}
	ms := delay / time.Millisecond
	if delay%time.Millisecond != 0 {
		ms++
	}
	bucket := ""
	if info.Bucket != "" {
		bucket = digest(info.Bucket + "\x00" + majorResource(path))
	}
	global := 0
	if info.Global {
		global = 1
	}
	opCtx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()
	_, err := l.eval(opCtx, updateScript, l.keys, routeID(method, path), bucket, info.Remaining, int64(ms), global, l.cacheTTL.Milliseconds(), l.globalLimit)
	if err != nil {
		l.remember(err)
		return fmt.Errorf("redislimiter: update: %w", err)
	}
	return nil
}
