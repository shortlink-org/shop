//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	osrmImage  = "registry.gitlab.com/shortlink-org/shop/osrm"
	osrmPort   = "5000/tcp"
	kafkaImage = "confluentinc/confluent-local:7.5.0"
)

// KafkaContainer holds a Kafka testcontainer and its broker list.
type KafkaContainer struct {
	kc      *kafka.KafkaContainer
	Brokers []string
}

// SetupKafkaContainer starts a Kafka container and returns brokers.
func SetupKafkaContainer(t *testing.T) *KafkaContainer {
	t.Helper()
	ctx := context.Background()

	kc, err := kafka.Run(ctx, kafkaImage)
	if err != nil {
		t.Fatalf("failed to start kafka container: %v", err)
	}

	brokers, err := kc.Brokers(ctx)
	if err != nil {
		_ = kc.Terminate(ctx)
		t.Fatalf("failed to get kafka brokers: %v", err)
	}

	c := &KafkaContainer{kc: kc, Brokers: brokers}
	t.Cleanup(func() {
		if err := kc.Terminate(ctx); err != nil {
			t.Logf("failed to terminate kafka container: %v", err)
		}
	})
	return c
}

// OSRMContainer holds an OSRM testcontainer and its base URL.
type OSRMContainer struct {
	Container testcontainers.Container
	BaseURL   string
}

// SetupOSRMContainer starts an OSRM container (Berlin) and returns the base URL.
func SetupOSRMContainer(t *testing.T) *OSRMContainer {
	t.Helper()
	ctx := context.Background()

	osrmC, err := testcontainers.Run(ctx, osrmImage,
		testcontainers.WithExposedPorts(osrmPort),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(osrmPort).WithStartupTimeout(120*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start osrm container: %v", err)
	}

	mappedPort, err := osrmC.MappedPort(ctx, "5000")
	if err != nil {
		_ = osrmC.Terminate(ctx)
		t.Fatalf("failed to get osrm mapped port: %v", err)
	}

	host, err := osrmC.Host(ctx)
	if err != nil {
		_ = osrmC.Terminate(ctx)
		t.Fatalf("failed to get osrm host: %v", err)
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	c := &OSRMContainer{Container: osrmC, BaseURL: baseURL}
	t.Cleanup(func() {
		if err := osrmC.Terminate(ctx); err != nil {
			t.Logf("failed to terminate osrm container: %v", err)
		}
	})
	return c
}
