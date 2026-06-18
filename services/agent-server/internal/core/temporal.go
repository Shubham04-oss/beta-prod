package core

import (
	"log"

	"github.com/synq/pkg/authcontext"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// TemporalManager wraps the Temporal SDK Client for dependency injection.
type TemporalManager struct {
	Client client.Client
}

// NewTemporalManager initializes the connection to the Temporal Cluster.
// In production, you would pass connection options (e.g., TLS certs, HostPort).
func NewTemporalManager(hostPort string) (*TemporalManager, error) {
	if hostPort == "" {
		hostPort = client.DefaultHostPort // localhost:7233
	}

	c, err := client.Dial(client.Options{
		HostPort:           hostPort,
		ContextPropagators: []workflow.ContextPropagator{authcontext.NewPillarContextPropagator()},
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to Temporal cluster at %s", hostPort)

	return &TemporalManager{
		Client: c,
	}, nil
}

// NewWorker creates a Temporal worker that listens to a specific task queue.
func (tm *TemporalManager) NewWorker(taskQueue string) worker.Worker {
	w := worker.New(tm.Client, taskQueue, worker.Options{})
	return w
}

// Close gracefully shuts down the client connection.
func (tm *TemporalManager) Close() {
	tm.Client.Close()
}
