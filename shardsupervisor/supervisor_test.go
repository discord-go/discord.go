package shardsupervisor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestChildHelper(t *testing.T) {
	if os.Getenv("SHARD_SUPERVISOR_HELPER") != "1" {
		return
	}
	data := os.Getenv("DISCORD_SHARD_ID") + "/" + os.Getenv("DISCORD_SHARD_COUNT")
	if err := os.WriteFile(os.Getenv("SHARD_SUPERVISOR_OUTPUT"), []byte(data), 0o600); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func TestSleepingChildHelper(t *testing.T) {
	if os.Getenv("SHARD_SUPERVISOR_SLEEP_HELPER") != "1" {
		return
	}
	time.Sleep(10 * time.Second)
	os.Exit(0)
}

func TestRunPassesShardAssignmentAndReportsExit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assignment")
	t.Setenv("SHARD_SUPERVISOR_HELPER", "1")
	t.Setenv("SHARD_SUPERVISOR_OUTPUT", path)
	t.Setenv("DISCORD_SHARD_ID", "wrong")
	err := Run(context.Background(), Config{Executable: os.Args[0], Args: []string{"-test.run=^TestChildHelper$"}, ShardCount: 4, ShardIDs: []int{2}})
	if err == nil || !strings.Contains(err.Error(), "shard 2 exited") {
		t.Fatalf("Run: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "2/4" {
		t.Fatalf("assignment %q: %v", data, err)
	}
}

func TestRunCancellationAndValidation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Run(ctx, Config{Executable: os.Args[0], ShardCount: 1, ShardIDs: []int{0}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Run: %v", err)
	}
	for _, cfg := range []Config{
		{Executable: os.Args[0], ShardCount: 1, ShardIDs: []int{0, 0}},
		{Executable: os.Args[0], ShardCount: 1, ShardIDs: []int{1}},
		{Executable: os.Args[0], ShardCount: 1, ShardIDs: []int{0}, StartInterval: time.Second},
		{Executable: os.Args[0], ShardCount: 1, ShardIDs: []int{0}, HealthInterval: -time.Second},
	} {
		if err := cfg.validate(); err == nil {
			t.Fatalf("accepted invalid config: %+v", cfg)
		}
	}
}

func TestFileCoordinatorAcquireRelease(t *testing.T) {
	coordinator := &FileCoordinator{Dir: t.TempDir()}
	release, err := coordinator.Acquire(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := coordinator.Acquire(ctx, 3); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second acquire: %v", err)
	}
	release()
	release2, err := coordinator.Acquire(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	release2()
}

func TestFileCoordinatorWaitIdentify(t *testing.T) {
	coordinator := &FileCoordinator{Dir: t.TempDir(), IdentifyGap: 100 * time.Millisecond}
	start := time.Now()
	if err := coordinator.WaitIdentify(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := coordinator.WaitIdentify(context.Background()); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Fatalf("identify gap = %v, want >= 100ms", elapsed)
	}
}

func TestFileCoordinatorRefresh(t *testing.T) {
	coordinator := &FileCoordinator{Dir: t.TempDir()}
	release, err := coordinator.Acquire(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if err := coordinator.Refresh(0); err != nil {
		t.Fatal(err)
	}
}

func TestRunHealthCheckKillsChild(t *testing.T) {
	t.Setenv("SHARD_SUPERVISOR_SLEEP_HELPER", "1")
	err := Run(context.Background(), Config{
		Executable:     os.Args[0],
		Args:           []string{"-test.run=^TestSleepingChildHelper$"},
		ShardCount:     1,
		ShardIDs:       []int{0},
		HealthCheck:    func(context.Context, int) error { return errors.New("unhealthy") },
		HealthInterval: 10 * time.Millisecond,
		HealthGrace:    10 * time.Millisecond,
	})
	if err == nil || !strings.Contains(err.Error(), "unhealthy") {
		t.Fatalf("Run: %v", err)
	}
}

type recordingCoordinator struct {
	mu       sync.Mutex
	acquired []int
	released []int
	waits    int
}

func (c *recordingCoordinator) Acquire(_ context.Context, shardID int) (func(), error) {
	c.mu.Lock()
	c.acquired = append(c.acquired, shardID)
	c.mu.Unlock()
	return func() {
		c.mu.Lock()
		c.released = append(c.released, shardID)
		c.mu.Unlock()
	}, nil
}

func (c *recordingCoordinator) WaitIdentify(context.Context) error {
	c.mu.Lock()
	c.waits++
	c.mu.Unlock()
	return nil
}

func TestRunUsesCoordinator(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assignment")
	t.Setenv("SHARD_SUPERVISOR_HELPER", "1")
	t.Setenv("SHARD_SUPERVISOR_OUTPUT", path)
	coordinator := &recordingCoordinator{}
	err := Run(context.Background(), Config{
		Executable:  os.Args[0],
		Args:        []string{"-test.run=^TestChildHelper$"},
		ShardCount:  2,
		ShardIDs:    []int{1},
		Coordinator: coordinator,
	})
	if err == nil {
		t.Fatal("expected child exit error")
	}
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if len(coordinator.acquired) != 1 || coordinator.acquired[0] != 1 {
		t.Fatalf("acquired = %v", coordinator.acquired)
	}
	if len(coordinator.released) != 1 || coordinator.released[0] != 1 {
		t.Fatalf("released = %v", coordinator.released)
	}
	if coordinator.waits != 1 {
		t.Fatalf("WaitIdentify calls = %d, want 1", coordinator.waits)
	}
}
