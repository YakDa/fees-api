package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
)

var (
	// Client is the global Temporal client
	Client client.Client
)

// TaskQueue is the default task queue for billing workflows
const TaskQueue = "billing"

// Initialize connects to the Temporal server
func Initialize(ctx context.Context, namespace, hostPort string) (client.Client, error) {
	if hostPort == "" {
		hostPort = "localhost:7233" // Default Temporal local address
	}

	var options client.Options
	if namespace != "" {
		options.Namespace = namespace
	}
	options.HostPort = hostPort

	c, err := client.Dial(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Temporal: %w", err)
	}

	Client = c
	return c, nil
}

// Close closes the Temporal client
func Close() {
	if Client != nil {
		Client.Close()
	}
}
