//go:build integration

package testhelpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer holds the container and client for tests.
type RedisContainer struct {
	Container testcontainers.Container
	Client    rueidis.Client
	Host      string
	Port      string
}

// SetupRedisContainer creates a Redis container for integration tests.
// It returns a RedisContainer with an active client and registers cleanup with t.Cleanup().
func SetupRedisContainer(t *testing.T) *RedisContainer {
	t.Helper()

	ctx := context.Background()

	container, err := redis.Run(ctx,
		"redis:8-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}

	// Get host and port
	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to get redis port: %v", err)
	}

	// Create rueidis client
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", host, port.Port())},
	})
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("failed to create redis client: %v", err)
	}

	// Verify connection
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		container.Terminate(ctx)
		t.Fatalf("failed to ping redis: %v", err)
	}

	rc := &RedisContainer{
		Container: container,
		Client:    client,
		Host:      host,
		Port:      port.Port(),
	}

	// Register cleanup
	t.Cleanup(func() {
		client.Close()
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	return rc
}

// FlushAll clears all data in Redis.
func (rc *RedisContainer) FlushAll(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	if err := rc.Client.Do(ctx, rc.Client.B().Flushall().Build()).Error(); err != nil {
		t.Fatalf("failed to flush redis: %v", err)
	}
}
