package client

import (
	"context"
)

// HealthInterface defines the health-related operations
type HealthInterface interface {
	Health(ctx context.Context) error
}

// HealthClient handles health-related requests
type HealthClient struct {
	client *BaseClient
}

// NewHealthClient creates a new health client
func NewHealthClient(client *BaseClient) HealthInterface {
	return &HealthClient{client: client}
}

// Health checks if the server is healthy
func (c *HealthClient) Health(ctx context.Context) error {
	resp, err := c.client.Get(ctx, "/health", "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
