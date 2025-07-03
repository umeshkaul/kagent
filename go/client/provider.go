package client

import (
	"context"
)

// ProviderInterface defines the provider operations
type ProviderInterface interface {
	ListSupportedModelProviders(ctx context.Context) ([]ProviderInfo, error)
	ListSupportedMemoryProviders(ctx context.Context) ([]ProviderInfo, error)
}

// ProviderClient handles provider-related requests
type ProviderClient struct {
	client *BaseClient
}

// NewProviderClient creates a new provider client
func NewProviderClient(client *BaseClient) ProviderInterface {
	return &ProviderClient{client: client}
}

// ListSupportedModelProviders lists all supported model providers
func (c *ProviderClient) ListSupportedModelProviders(ctx context.Context) ([]ProviderInfo, error) {
	resp, err := c.client.Get(ctx, "/api/providers/models", "")
	if err != nil {
		return nil, err
	}

	var providers []ProviderInfo
	if err := DecodeResponse(resp, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

// ListSupportedMemoryProviders lists all supported memory providers
func (c *ProviderClient) ListSupportedMemoryProviders(ctx context.Context) ([]ProviderInfo, error) {
	resp, err := c.client.Get(ctx, "/api/providers/memories", "")
	if err != nil {
		return nil, err
	}

	var providers []ProviderInfo
	if err := DecodeResponse(resp, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}
