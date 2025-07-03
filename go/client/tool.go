package client

import (
	"context"
	"fmt"
)

// ToolInterface defines the tool operations
type ToolInterface interface {
	ListTools(ctx context.Context, userID string) ([]Tool, error)
}

// ToolClient handles tool-related requests
type ToolClient struct {
	client *BaseClient
}

// NewToolClient creates a new tool client
func NewToolClient(client *BaseClient) ToolInterface {
	return &ToolClient{client: client}
}

// ListTools lists all tools for a user
func (c *ToolClient) ListTools(ctx context.Context, userID string) ([]Tool, error) {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.client.Get(ctx, "/api/tools", userID)
	if err != nil {
		return nil, err
	}

	var tools []Tool
	if err := DecodeResponse(resp, &tools); err != nil {
		return nil, err
	}

	return tools, nil
}
