package client

import (
	"context"
)

// ModelInterface defines the model operations
type ModelInterface interface {
	ListSupportedModels(ctx context.Context) (interface{}, error)
}

// ModelClient handles model-related requests
type ModelClient struct {
	client *BaseClient
}

// NewModelClient creates a new model client
func NewModelClient(client *BaseClient) ModelInterface {
	return &ModelClient{client: client}
}

// ListSupportedModels lists all supported models
func (c *ModelClient) ListSupportedModels(ctx context.Context) (interface{}, error) {
	resp, err := c.client.Get(ctx, "/api/models", "")
	if err != nil {
		return nil, err
	}

	var models interface{}
	if err := DecodeResponse(resp, &models); err != nil {
		return nil, err
	}

	return models, nil
}
