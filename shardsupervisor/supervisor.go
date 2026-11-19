// Package shardsupervisor runs one child process for each assigned Discord
// gateway shard. Children own their connections and must exit when unhealthy.
package shardsupervisor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

const minimumStartInterval = 5 * time.Second

// Config describes static ownership within one supervisor. A Coordinator can
// serialize shard ownership and Identify starts across supervisors.
type Config struct {
	Executable string
	Args       []string
	ShardCount int
	ShardIDs   []int
	// MaxRestarts is the number of retries after a child exits. Zero disables retries.
	MaxRestarts int
	// RestartDelay is the initial delay between child exits and retries.
	RestartDelay time.Duration
	// StartInterval spaces child starts within this supervisor; at least 5s.
	StartInterval time.Duration
	// Coordinator, when set, claims each shard before its child starts and
	// gates Identify starts across supervisors. See FileCoordinator.
	Coordinator Coordinator
	// HealthCheck, when set, runs after HealthGrace and then every
	// HealthInterval. A non-nil error kills the child so the restart policy
	// applies. The callback must return when its context is done.
	HealthCheck func(ctx context.Context, shardID int) error
	// HealthInterval is the delay between health checks. Defaults to 10s.
	HealthInterval time.Duration
	// HealthGrace is how long a child may run before the first health check.
	// Defaults to 30s.
	HealthGrace time.Duration
}

func (c Config) validate() error {
	if c.Executable == "" || c.ShardCount <= 0 || len(c.ShardIDs) == 0 {
		return errors.New("shardsupervisor: executable, shard count and shard IDs are required")
	}
	if c.MaxRestarts < 0 || c.RestartDelay < 0 || c.StartInterval < 0 || (c.StartInterval > 0 && c.StartInterval < minimumStartInterval) {
		return errors.New("shardsupervisor: invalid restart policy or start interval")
	}
	if c.HealthInterval < 0 || c.HealthGrace < 0 {
		return errors.New("shardsupervisor: invalid health interval or grace period")
	}
	seen := make(map[int]bool, len(c.ShardIDs))
	for _, id := range c.ShardIDs {
		if id < 0 || id >= c.ShardCount || seen[id] {
			return fmt.Errorf("shardsupervisor: invalid or duplicate shard ID %d", id)
		}
		seen[id] = true
	}
	return nil
}

type starter struct {
	mu       sync.Mutex
	next     time.Time
	interval time.Duration
}

func (s *starter) wait(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if delay := time.Until(s.next); delay > 0 {
		t := time.NewTimer(delay)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.next = time.Now().Add(s.interval)
	return nil
}

// Run starts the configured children and waits. Any child exit counts as a
// failure, including exit status zero. After its restart budget is exhausted,
// all other children are stopped and Run reports an error. Cancellation kills
// children and returns ctx.Err(). Child processes inherit standard output,
// standard error and the environment, with DISCORD_SHARD_ID and
// DISCORD_SHARD_COUNT set for each child.
//
// With a Coordinator, Run claims each shard before starting its child and
// releases the claim when the child exits. It also calls WaitIdentify before
// each start; children that identify later should call WaitIdentify
// themselves for exact spacing. With HealthCheck, Run kills a child whose
// check fails so the restart policy applies.
func Run(ctx context.Context, cfg Config) error {
	if ctx == nil {
		return errors.New("shardsupervisor: nil context")
	}
	if err := cfg.validate(); err != nil {
		return err
	}
	if cfg.StartInterval == 0 {
		cfg.StartInterval = minimumStartInterval
	}
	if cfg.RestartDelay == 0 {
		cfg.RestartDelay = time.Second
	}
	if cfg.HealthInterval == 0 {
		cfg.HealthInterval = 10 * time.Second
	}
	if cfg.HealthGrace == 0 {
		cfg.HealthGrace = 30 * time.Second
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	gate := &starter{interval: cfg.StartInterval}
	results := make(chan error, len(cfg.ShardIDs))
	for _, id := range cfg.ShardIDs {
		id := id
		go func() {
			for attempt := 0; ; attempt++ {
				if cfg.Coordinator != nil {
					if err := cfg.Coordinator.WaitIdentify(ctx); err != nil {
						results <- err
						return
					}
				}
				if err := gate.wait(ctx); err != nil {
					results <- err
					return
				}
				var release func()
				if cfg.Coordinator != nil {
					var err error
					release, err = cfg.Coordinator.Acquire(ctx, id)
					if err != nil {
						results <- fmt.Errorf("shardsupervisor: shard %d claim: %w", id, err)
						return
					}
				}
				cmd := exec.CommandContext(ctx, cfg.Executable, cfg.Args...)
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				cmd.Env = childEnv(os.Environ(), id, cfg.ShardCount)
				if err := cmd.Start(); err != nil {
					if release != nil {
						release()
					}
					results <- fmt.Errorf("shardsupervisor: shard %d start: %w", id, err)
					return
				}
				healthCtx, healthCancel := context.WithCancel(ctx)
				healthErr := make(chan error, 1)
				go func() { healthErr <- runHealth(healthCtx, cfg, id, cmd) }()
				err := cmd.Wait()
				healthCancel()
				healthFailure := <-healthErr
				if release != nil {
					release()
				}
				if ctx.Err() != nil {
					results <- ctx.Err()
					return
				}
				if healthFailure != nil {
					err = healthFailure
				}
				if attempt >= cfg.MaxRestarts {
					if err == nil {
						err = errors.New("exited successfully")
					}
					results <- fmt.Errorf("shardsupervisor: shard %d exited after %d restarts: %w", id, attempt, err)
					return
				}
				t := time.NewTimer(cfg.RestartDelay)
				select {
				case <-ctx.Done():
					t.Stop()
					results <- ctx.Err()
					return
				case <-t.C:
				}
			}
		}()
	}
	first := <-results
	cancel()
	for i := 1; i < len(cfg.ShardIDs); i++ {
		<-results
	}
	return first
}

func runHealth(ctx context.Context, cfg Config, shardID int, cmd *exec.Cmd) error {
	refresher, _ := cfg.Coordinator.(Refresher)
	if cfg.HealthCheck == nil && refresher == nil {
		return nil
	}
	if cfg.HealthGrace > 0 {
		timer := time.NewTimer(cfg.HealthGrace)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
	ticker := time.NewTicker(cfg.HealthInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if cfg.HealthCheck != nil {
				checkCtx, checkCancel := context.WithTimeout(ctx, cfg.HealthInterval)
				err := cfg.HealthCheck(checkCtx, shardID)
				checkCancel()
				if err != nil {
					_ = cmd.Process.Kill()
					return err
				}
			}
			if refresher != nil {
				_ = refresher.Refresh(shardID)
			}
		}
	}
}

func childEnv(env []string, id, count int) []string {
	result := make([]string, 0, len(env)+2)
	for _, item := range env {
		if !strings.HasPrefix(item, "DISCORD_SHARD_ID=") && !strings.HasPrefix(item, "DISCORD_SHARD_COUNT=") {
			result = append(result, item)
		}
	}
	return append(result, "DISCORD_SHARD_ID="+strconv.Itoa(id), "DISCORD_SHARD_COUNT="+strconv.Itoa(count))
}
