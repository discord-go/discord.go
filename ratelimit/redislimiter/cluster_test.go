package redislimiter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/discord-go/discord.go/ratelimit"
)

// testRedisCluster starts three isolated loopback-only Redis nodes. Each EVAL
// uses redis-cli's cluster routing so connections to different nodes can reach
// the key's owner. No existing Redis service or cluster is modified.
func testRedisCluster(t *testing.T) func(int) EvalFunc {
	t.Helper()
	server, err := exec.LookPath("redis-server")
	if err != nil {
		t.Skip("redis-server is not installed")
	}
	cli, err := exec.LookPath("redis-cli")
	if err != nil {
		t.Skip("redis-cli is not installed")
	}

	// Redis Cluster uses a second TCP port, 10000 above the client port.
	// Reserve both ports for all nodes until immediately before startup.
	type pair struct {
		port        int
		client, bus net.Listener
	}
	var reserved []pair
	defer func() {
		for _, p := range reserved {
			_ = p.client.Close()
			_ = p.bus.Close()
		}
	}()
	for attempts := 0; len(reserved) < 3 && attempts < 100; attempts++ {
		listener, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		port := listener.Addr().(*net.TCPAddr).Port
		if port > 55535 {
			listener.Close()
			continue
		}
		bus, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", strconv.Itoa(port+10000)))
		if err != nil {
			listener.Close()
			continue
		}
		reserved = append(reserved, pair{port: port, client: listener, bus: bus})
	}
	if len(reserved) != 3 {
		t.Fatal("unable to reserve Redis Cluster ports")
	}
	root := t.TempDir()
	var processes []*exec.Cmd
	t.Cleanup(func() {
		for _, cmd := range processes {
			_ = cmd.Process.Kill()
		}
		for _, cmd := range processes {
			_ = cmd.Wait()
		}
	})
	for i, p := range reserved {
		dir := filepath.Join(root, strconv.Itoa(i))
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		_ = p.client.Close()
		_ = p.bus.Close()
		cmd := exec.Command(server,
			"--bind", "127.0.0.1", "--port", strconv.Itoa(p.port),
			"--protected-mode", "yes", "--save", "", "--appendonly", "no",
			"--cluster-enabled", "yes", "--cluster-node-timeout", "1000",
			"--cluster-config-file", filepath.Join(dir, "nodes.conf"),
			"--dir", dir, "--logfile", filepath.Join(dir, "redis.log"))
		if err := cmd.Start(); err != nil {
			t.Fatalf("start Redis node %d: %v", i, err)
		}
		processes = append(processes, cmd)
		deadline := time.Now().Add(8 * time.Second)
		for {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p.port)), 100*time.Millisecond)
			if err == nil {
				conn.Close()
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("Redis node %d did not start", i)
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	args := []string{"--cluster", "create"}
	for _, p := range reserved {
		args = append(args, net.JoinHostPort("127.0.0.1", strconv.Itoa(p.port)))
	}
	args = append(args, "--cluster-replicas", "0", "--cluster-yes")
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, cli, args...).CombinedOutput(); err != nil {
		t.Fatalf("create isolated Redis Cluster: %v: %s", err, out)
	}
	// Slot assignment can finish before every node reports a healthy cluster.
	for i, p := range reserved {
		deadline := time.Now().Add(15 * time.Second)
		for {
			out, err := exec.CommandContext(ctx, cli, "-h", "127.0.0.1", "-p", strconv.Itoa(p.port), "CLUSTER", "INFO").Output()
			if err == nil && strings.Contains(string(out), "cluster_state:ok") {
				break
			}
			if time.Now().After(deadline) || ctx.Err() != nil {
				t.Fatalf("Redis Cluster node %d did not become healthy: %v", i, err)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	return func(node int) EvalFunc {
		return func(ctx context.Context, script string, keys []string, args ...any) (any, error) {
			argv := []string{"-h", "127.0.0.1", "-p", strconv.Itoa(reserved[node].port), "-c", "-e", "--json", "EVAL", script, strconv.Itoa(len(keys))}
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
}

func TestRedisClusterCoordinatesWorkersAcrossNodes(t *testing.T) {
	eval := testRedisCluster(t)
	a, err := New(eval(0), Options{Namespace: "cluster-test"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := New(eval(1), Options{Namespace: "cluster-test"})
	if err != nil {
		t.Fatal(err)
	}
	path := "/channels/123/messages"
	first, err := a.AcquireRoute(requestContext(t), "GET", path)
	if err != nil {
		t.Fatal(err)
	}
	defer first()
	blocked, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	if release, err := b.AcquireRoute(blocked, "GET", path); err == nil {
		release()
		t.Fatal("second node admitted an overlapping discovery request")
	} else if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error while waiting: %v", err)
	}
	if err := a.UpdateRouteContext(context.Background(), "GET", path, ratelimit.Info{Bucket: "shared", Remaining: 1, ResetAfter: 10 * time.Second}); err != nil {
		t.Fatal(err)
	}
	first()
	second, err := b.AcquireRoute(requestContext(t), "GET", path)
	if err != nil {
		t.Fatal(err)
	}
	second()
	blocked2, cancel2 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel2()
	if release, err := a.AcquireRoute(blocked2, "GET", path); err == nil {
		release()
		t.Fatal("cluster admitted more requests than the shared bucket permits")
	} else if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error while waiting for shared bucket: %v", err)
	}
}
