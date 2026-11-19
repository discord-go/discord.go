// The same executable runs as supervisor or as one assigned shard worker.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/shardsupervisor"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	coordDir := os.Getenv("DISCORD_COORDINATION_DIR")
	if rawID := os.Getenv("DISCORD_SHARD_ID"); rawID != "" {
		id, err := strconv.Atoi(rawID)
		if err != nil {
			log.Fatal(err)
		}
		count, err := strconv.Atoi(os.Getenv("DISCORD_SHARD_COUNT"))
		if err != nil {
			log.Fatal(err)
		}
		if coordDir != "" {
			coordinator := &shardsupervisor.FileCoordinator{Dir: coordDir}
			if err := coordinator.WaitIdentify(ctx); err != nil {
				log.Fatal(err)
			}
			go heartbeat(ctx, filepath.Join(coordDir, fmt.Sprintf("health-%d", id)))
		}
		client := bot.New(os.Getenv("DISCORD_TOKEN"), bot.WithIntents(intents.Guilds),
			bot.WithShards(count), bot.WithShardIDs(id), bot.WithCommandSyncDisabled())
		if err := client.RunContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatal(err)
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	// Select a shard count and ownership based on your Discord deployment.
	cfg := shardsupervisor.Config{
		Executable: executable, ShardCount: 2, ShardIDs: []int{0, 1},
		MaxRestarts: 3, RestartDelay: 10 * time.Second,
	}
	if coordDir != "" {
		cfg.Coordinator = &shardsupervisor.FileCoordinator{Dir: coordDir}
		cfg.HealthCheck = func(_ context.Context, shardID int) error {
			path := filepath.Join(coordDir, fmt.Sprintf("health-%d", shardID))
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			if time.Since(info.ModTime()) > 15*time.Second {
				return fmt.Errorf("shard %d heartbeat is stale", shardID)
			}
			return nil
		}
		cfg.HealthGrace = 30 * time.Second
		cfg.HealthInterval = 10 * time.Second
	}
	if err := shardsupervisor.Run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func heartbeat(ctx context.Context, path string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		_ = os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339Nano)), 0o600)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
