package shardsupervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Coordinator serializes shard ownership and Identify starts across
// supervisors. Implementations must be safe for concurrent use.
type Coordinator interface {
	// Acquire claims a shard. It blocks until the shard is free or ctx is
	// done. The returned release function must be called when the child exits.
	Acquire(ctx context.Context, shardID int) (release func(), err error)
	// WaitIdentify blocks until this process may start the next Identify.
	WaitIdentify(ctx context.Context) error
}

// Refresher is implemented by coordinators that can keep a shard claim alive
// while its child is running. The supervisor calls Refresh from its health
// loop so a long-running child is not mistaken for a stale claim.
type Refresher interface {
	Refresh(shardID int) error
}

// FileCoordinator coordinates supervisors that share a directory. Shard
// claims are O_CREATE|O_EXCL lock files; Identify starts are serialized by a
// timestamp file and a short-lived lock. Stale claims older than StaleAfter
// are removed, so a crashed supervisor does not block a shard forever.
type FileCoordinator struct {
	Dir string
	// IdentifyGap is the minimum time between Identify starts across all
	// supervisors sharing Dir. Defaults to 5s.
	IdentifyGap time.Duration
	// StaleAfter is how old a lock file may be before it is treated as stale.
	// Defaults to 2m. The supervisor refreshes live claims from its health
	// loop.
	StaleAfter time.Duration
}

func (c *FileCoordinator) identifyGap() time.Duration {
	if c.IdentifyGap > 0 {
		return c.IdentifyGap
	}
	return 5 * time.Second
}

func (c *FileCoordinator) staleAfter() time.Duration {
	if c.StaleAfter > 0 {
		return c.StaleAfter
	}
	return 2 * time.Minute
}

func (c *FileCoordinator) ensureDir() error {
	if c.Dir == "" {
		return errors.New("shardsupervisor: coordinator directory is required")
	}
	return os.MkdirAll(c.Dir, 0o700)
}

func (c *FileCoordinator) shardPath(shardID int) string {
	return filepath.Join(c.Dir, fmt.Sprintf("shard-%d.lock", shardID))
}

// Acquire claims shardID. A second supervisor sharing Dir waits until the
// first releases the claim or the claim becomes stale.
func (c *FileCoordinator) Acquire(ctx context.Context, shardID int) (func(), error) {
	if err := c.ensureDir(); err != nil {
		return nil, err
	}
	return acquireLock(ctx, c.shardPath(shardID), c.staleAfter())
}

// Refresh updates the claim's modification time.
func (c *FileCoordinator) Refresh(shardID int) error {
	if err := c.ensureDir(); err != nil {
		return err
	}
	now := time.Now()
	return os.Chtimes(c.shardPath(shardID), now, now)
}

// WaitIdentify serializes Identify starts across supervisors sharing Dir.
func (c *FileCoordinator) WaitIdentify(ctx context.Context) error {
	if err := c.ensureDir(); err != nil {
		return err
	}
	unlock, err := acquireLock(ctx, filepath.Join(c.Dir, "identify.lock"), c.staleAfter())
	if err != nil {
		return err
	}
	defer unlock()

	timestampPath := filepath.Join(c.Dir, "identify.timestamp")
	var last time.Time
	if data, readErr := os.ReadFile(timestampPath); readErr == nil {
		if parsed, parseErr := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data))); parseErr == nil {
			last = parsed
		}
	}
	if wait := time.Until(last.Add(c.identifyGap())); wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	return os.WriteFile(timestampPath, []byte(time.Now().UTC().Format(time.RFC3339Nano)), 0o600)
}

func acquireLock(ctx context.Context, path string, staleAfter time.Duration) (func(), error) {
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = file.Close()
			var once sync.Once
			return func() { once.Do(func() { _ = os.Remove(path) }) }, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		if info, statErr := os.Stat(path); statErr == nil && staleAfter > 0 && time.Since(info.ModTime()) > staleAfter {
			_ = os.Remove(path)
			continue
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
