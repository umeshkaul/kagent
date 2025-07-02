package database

import (
	"fmt"
	"slices"

	autogen_client "github.com/kagent-dev/kagent/go/internal/autogen/client"
)

type Client interface {
	CreateFeedback(feedback *Feedback) error
	CreateRun(req *Run) error
	CreateSession(session *Session) error
	CreateTeam(team *Team) error
	UpsertTeam(team *Team) error
	CreateToolServer(toolServer *ToolServer) (*ToolServer, error)
	RefreshToolsForServer(serverName string, tools []*autogen_client.NamedTool) error
	DeleteRun(runID int) error
	DeleteSession(sessionName string, userID string) error
	DeleteTeam(teamName string) error
	DeleteToolServer(serverName string) error
	GetRun(runID int) (*Run, error)
	GetRunMessages(runID int) ([]Message, error)
	GetSession(name string, userID string) (*Session, error)
	GetTeam(name string) (*Team, error)
	GetTool(name string) (*Tool, error)
	GetToolServer(name string) (*ToolServer, error)
	ListFeedback(userID string) ([]Feedback, error)
	ListRuns(userID string) ([]Run, error)
	ListSessionRuns(sessionName string, userID string) ([]Run, error)
	ListSessions(userID string) ([]Session, error)
	ListTeams(userID string) ([]Team, error)
	ListToolServers() ([]ToolServer, error)
	CreateTool(tool *Tool) error
	ListTools(userID string) ([]Tool, error)
	ListToolsForServer(serverName string) ([]Tool, error)
	ListMessagesForRun(runID uint) ([]Message, error)
	UpdateSession(session *Session) error
	UpdateToolServer(server *ToolServer) error
	UpdateRun(run *Run) error
	UpdateTeam(team *Team) error
}

type clientImpl struct {
	serviceWrapper *ServiceWrapper
}

func NewClient(serviceWrapper *ServiceWrapper) Client {
	return &clientImpl{
		serviceWrapper: serviceWrapper,
	}
}

// CreateFeedback creates a new feedback record
func (c *clientImpl) CreateFeedback(feedback *Feedback) error {
	return c.serviceWrapper.Feedback.Create(feedback)
}

// CreateRun creates a new run record
func (c *clientImpl) CreateRun(req *Run) error {
	return c.serviceWrapper.Run.Create(req)
}

// CreateSession creates a new session record
func (c *clientImpl) CreateSession(session *Session) error {
	return c.serviceWrapper.Session.Create(session)
}

// CreateTeam creates a new team record
func (c *clientImpl) CreateTeam(team *Team) error {
	return c.serviceWrapper.Team.Create(team)
}

// UpsertTeam upserts a team record
func (c *clientImpl) UpsertTeam(team *Team) error {
	return c.serviceWrapper.Team.Update(team)
}

// CreateToolServer creates a new tool server record
func (c *clientImpl) CreateToolServer(toolServer *ToolServer) (*ToolServer, error) {
	err := c.serviceWrapper.ToolServer.Create(toolServer)
	if err != nil {
		return nil, err
	}
	return toolServer, nil
}

// CreateTool creates a new tool record
func (c *clientImpl) CreateTool(tool *Tool) error {
	return c.serviceWrapper.Tool.Create(tool)
}

// DeleteRun deletes a run by ID
func (c *clientImpl) DeleteRun(runID int) error {
	return c.serviceWrapper.Run.Delete(Clause{Key: "id", Value: runID})
}

// DeleteSession deletes a session by name and user ID
func (c *clientImpl) DeleteSession(sessionName string, userID string) error {
	return c.serviceWrapper.Session.Delete(
		Clause{Key: "name", Value: sessionName},
		Clause{Key: "user_id", Value: userID},
	)
}

// DeleteTeam deletes a team by name and user ID
func (c *clientImpl) DeleteTeam(teamName string) error {
	return c.serviceWrapper.Team.Delete(Clause{Key: "name", Value: teamName})
}

// DeleteToolServer deletes a tool server by name and user ID
func (c *clientImpl) DeleteToolServer(serverName string) error {
	return c.serviceWrapper.ToolServer.Delete(Clause{Key: "name", Value: serverName})
}

// GetRun retrieves a run by ID
func (c *clientImpl) GetRun(runID int) (*Run, error) {
	return c.serviceWrapper.Run.Get(Clause{Key: "id", Value: runID})
}

// GetRunMessages retrieves messages for a specific run
func (c *clientImpl) GetRunMessages(runID int) ([]Message, error) {
	messages, err := c.serviceWrapper.Message.List(Clause{Key: "run_id", Value: runID})
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetSession retrieves a session by name and user ID
func (c *clientImpl) GetSession(sessionLabel string, userID string) (*Session, error) {
	return c.serviceWrapper.Session.Get(
		Clause{Key: "name", Value: sessionLabel},
		Clause{Key: "user_id", Value: userID},
	)
}

// GetTeam retrieves a team by name and user ID
func (c *clientImpl) GetTeam(teamLabel string) (*Team, error) {
	return c.serviceWrapper.Team.Get(Clause{Key: "name", Value: teamLabel})
}

// GetTool retrieves a tool by provider (name) and user ID
func (c *clientImpl) GetTool(provider string) (*Tool, error) {
	return c.serviceWrapper.Tool.Get(Clause{Key: "name", Value: provider})
}

// GetToolServer retrieves a tool server by name and user ID
func (c *clientImpl) GetToolServer(serverName string) (*ToolServer, error) {
	return c.serviceWrapper.ToolServer.Get(Clause{Key: "name", Value: serverName})
}

// ListFeedback lists all feedback for a user
func (c *clientImpl) ListFeedback(userID string) ([]Feedback, error) {
	feedback, err := c.serviceWrapper.Feedback.List(Clause{Key: "user_id", Value: userID})
	if err != nil {
		return nil, err
	}

	return feedback, nil
}

// ListRuns lists all runs for a user
func (c *clientImpl) ListRuns(userID string) ([]Run, error) {
	runs, err := c.serviceWrapper.Run.List(Clause{Key: "user_id", Value: userID})
	if err != nil {
		return nil, err
	}

	return runs, nil
}

// ListSessionRuns lists all runs for a specific session
func (c *clientImpl) ListSessionRuns(sessionName string, userID string) ([]Run, error) {
	return c.serviceWrapper.Run.List(
		Clause{Key: "session_name", Value: sessionName},
		Clause{Key: "user_id", Value: userID},
	)
}

// ListSessions lists all sessions for a user
func (c *clientImpl) ListSessions(userID string) ([]Session, error) {
	return c.serviceWrapper.Session.List(Clause{Key: "user_id", Value: userID})
}

// ListTeams lists all teams for a user
func (c *clientImpl) ListTeams(userID string) ([]Team, error) {
	return c.serviceWrapper.Team.List(Clause{Key: "user_id", Value: userID})
}

// ListToolServers lists all tool servers for a user
func (c *clientImpl) ListToolServers() ([]ToolServer, error) {
	return c.serviceWrapper.ToolServer.List()
}

// ListTools lists all tools for a user
func (c *clientImpl) ListTools(userID string) ([]Tool, error) {
	return c.serviceWrapper.Tool.List(Clause{Key: "user_id", Value: userID})
}

// ListToolsForServer lists all tools for a specific server
func (c *clientImpl) ListToolsForServer(serverName string) ([]Tool, error) {
	return c.serviceWrapper.Tool.List(Clause{Key: "server_name", Value: serverName})
}

// RefreshToolsForServer refreshes a tool server
func (c *clientImpl) RefreshToolsForServer(serverName string, tools []*autogen_client.NamedTool) error {
	existingTools, err := c.ListToolsForServer(serverName)
	if err != nil {
		return err
	}

	// Check if the tool exists in the existing tools
	// If it does, update it
	// If it doesn't, create it
	// If it's in the existing tools but not in the new tools, delete it
	for _, tool := range tools {
		existingToolIndex := slices.IndexFunc(existingTools, func(t Tool) bool {
			return t.Name == tool.Name
		})
		if existingToolIndex != -1 {
			existingTool := existingTools[existingToolIndex]
			existingTool.Component = *tool.Component
			err = c.serviceWrapper.Tool.Update(&existingTool)
			if err != nil {
				return err
			}
		} else {
			err = c.serviceWrapper.Tool.Create(&Tool{
				Name:      tool.Name,
				Component: *tool.Component,
			})
			if err != nil {
				return fmt.Errorf("failed to create tool %s: %v", tool.Name, err)
			}
		}
	}

	// Delete any tools that are in the existing tools but not in the new tools
	for _, existingTool := range existingTools {
		if !slices.ContainsFunc(tools, func(t *autogen_client.NamedTool) bool {
			return t.Name == existingTool.Name
		}) {
			err = c.serviceWrapper.Tool.Delete(Clause{Key: "name", Value: existingTool.Name})
			if err != nil {
				return fmt.Errorf("failed to delete tool %s: %v", existingTool.Name, err)
			}
		}
	}
	return nil
}

// UpdateSession updates a session
func (c *clientImpl) UpdateSession(session *Session) error {
	return c.serviceWrapper.Session.Update(session)
}

// UpdateToolServer updates a tool server
func (c *clientImpl) UpdateToolServer(server *ToolServer) error {
	return c.serviceWrapper.ToolServer.Update(server)
}

// UpdateRun updates a run record
func (c *clientImpl) UpdateRun(run *Run) error {
	return c.serviceWrapper.Run.Update(run)
}

// UpdateTeam updates a team record
func (c *clientImpl) UpdateTeam(team *Team) error {
	return c.serviceWrapper.Team.Update(team)
}

// ListMessagesForRun retrieves messages for a specific run (helper method)
func (c *clientImpl) ListMessagesForRun(runID uint) ([]Message, error) {
	return c.serviceWrapper.Message.List(Clause{Key: "run_id", Value: runID})
}
