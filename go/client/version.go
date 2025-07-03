package client

import (
	"context"
)

// VersionInterface defines the version-related operations
type VersionInterface interface {
	GetVersion(ctx context.Context) (*VersionResponse, error)
}

// VersionClient handles version-related requests
type VersionClient struct {
	client *BaseClient
}

// NewVersionClient creates a new version client
func NewVersionClient(client *BaseClient) VersionInterface {
	return &VersionClient{client: client}
}

// GetVersion retrieves version information
func (c *VersionClient) GetVersion(ctx context.Context) (*VersionResponse, error) {
	resp, err := c.client.Get(ctx, "/version", "")
	if err != nil {
		return nil, err
	}

	var version VersionResponse
	if err := DecodeResponse(resp, &version); err != nil {
		return nil, err
	}

	return &version, nil
}
