package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kagent-dev/kagent/go/client/api"
)

// TeamInterface defines the team operations
type TeamInterface interface {
	ListTeams(ctx context.Context, userID string) ([]Team, error)
	CreateTeam(ctx context.Context, request *api.TeamRequest) (*Team, error)
	GetTeam(ctx context.Context, teamID string) (*Team, error)
	UpdateTeam(ctx context.Context, teamID string, request *api.TeamRequest) (*Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
}

// TeamClient handles team-related requests
type TeamClient struct {
	client *BaseClient
}

// NewTeamClient creates a new team client
func NewTeamClient(client *BaseClient) TeamInterface {
	return &TeamClient{client: client}
}

// ListTeams lists all teams for a user
func (c *TeamClient) ListTeams(ctx context.Context, userID string) ([]Team, error) {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.client.Get(ctx, "/api/teams", userID)
	if err != nil {
		return nil, err
	}

	var response StandardResponse[[]Team]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var teams []Team
	if err := json.Unmarshal(teamsData, &teams); err != nil {
		return nil, err
	}

	return teams, nil
}

// CreateTeam creates a new team
func (c *TeamClient) CreateTeam(ctx context.Context, request *TeamRequest) (*Team, error) {
	resp, err := c.client.Post(ctx, "/api/teams", request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Team]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(teamData, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// GetTeam retrieves a specific team
func (c *TeamClient) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	path := fmt.Sprintf("/api/teams/%s", teamID)
	resp, err := c.client.Get(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Team]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(teamData, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// UpdateTeam updates an existing team
func (c *TeamClient) UpdateTeam(ctx context.Context, teamID string, request *TeamRequest) (*Team, error) {
	path := fmt.Sprintf("/api/teams/%s", teamID)
	resp, err := c.client.Put(ctx, path, request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Team]
	if err := DecodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(teamData, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// DeleteTeam deletes a team
func (c *TeamClient) DeleteTeam(ctx context.Context, teamID string) error {
	path := fmt.Sprintf("/api/teams/%s", teamID)
	resp, err := c.client.Delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
