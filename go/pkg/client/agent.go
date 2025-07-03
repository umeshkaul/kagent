package client

import (
	"context"
	"fmt"

	"github.com/kagent-dev/kagent/go/pkg/client/api"
)

// Agent defines the agent operations
type Agent interface {
	ListAgents(ctx context.Context, userID string) (*api.StandardResponse[[]api.Agent], error)
	CreateAgent(ctx context.Context, request *api.AgentRequest) (*api.StandardResponse[*api.Agent], error)
	GetAgent(ctx context.Context, agentRef string) (*api.StandardResponse[*api.Agent], error)
	UpdateAgent(ctx context.Context, request *api.AgentRequest) (*api.StandardResponse[*api.Agent], error)
	DeleteAgent(ctx context.Context, agentRef string) error
}

// teamClient handles team-related requests
type teamClient struct {
	client *BaseClient
}

// NewTeamClient creates a new team client
func NewTeamClient(client *BaseClient) Agent {
	return &teamClient{client: client}
}

// ListTeams lists all teams for a user
func (c *teamClient) ListAgents(ctx context.Context, userID string) (*api.StandardResponse[[]api.Agent], error) {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.client.Get(ctx, "/api/agents", userID)
	if err != nil {
		return nil, err
	}

	var response api.StandardResponse[[]api.Agent]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateTeam creates a new team
func (c *teamClient) CreateAgent(ctx context.Context, request *api.AgentRequest) (*api.StandardResponse[*api.Agent], error) {
	resp, err := c.client.Post(ctx, "/api/agents", request, "")
	if err != nil {
		return nil, err
	}

	var response api.StandardResponse[*api.Agent]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetTeam retrieves a specific team
func (c *teamClient) GetAgent(ctx context.Context, agentRef string) (*api.StandardResponse[*api.Agent], error) {
	path := fmt.Sprintf("/api/agents/%s", agentRef)
	resp, err := c.client.Get(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var response api.StandardResponse[*api.Agent]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// UpdateTeam updates an existing team
func (c *teamClient) UpdateAgent(ctx context.Context, request *api.AgentRequest) (*api.StandardResponse[*api.Agent], error) {
	path := fmt.Sprintf("/api/agents/%s", request.AgentRef)
	resp, err := c.client.Put(ctx, path, request, "")
	if err != nil {
		return nil, err
	}

	var response api.StandardResponse[*api.Agent]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteTeam deletes a team
func (c *teamClient) DeleteAgent(ctx context.Context, agentRef string) error {
	path := fmt.Sprintf("/api/agents/%s", agentRef)
	resp, err := c.client.Delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
