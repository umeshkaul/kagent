package client

import (
	"context"
)

// NamespaceInterface defines the namespace operations
type NamespaceInterface interface {
	ListNamespaces(ctx context.Context) ([]NamespaceResponse, error)
}

// NamespaceClient handles namespace-related requests
type NamespaceClient struct {
	client *BaseClient
}

// NewNamespaceClient creates a new namespace client
func NewNamespaceClient(client *BaseClient) NamespaceInterface {
	return &NamespaceClient{client: client}
}

// ListNamespaces lists all namespaces
func (c *NamespaceClient) ListNamespaces(ctx context.Context) ([]NamespaceResponse, error) {
	resp, err := c.client.Get(ctx, "/api/namespaces", "")
	if err != nil {
		return nil, err
	}

	var namespaces []NamespaceResponse
	if err := DecodeResponse(resp, &namespaces); err != nil {
		return nil, err
	}

	return namespaces, nil
}
