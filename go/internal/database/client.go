package database

import (
	"fmt"
	"slices"

	"github.com/kagent-dev/kagent/go/internal/autogen/api"
)

type Client interface {
	CreateFeedback(feedback *Feedback) error
	CreateRun(req *Run) error
	CreateSession(session *Session) error
	CreateTeam(team *Team) error
	UpsertTeam(team *Team) error
	CreateToolServer(toolServer *ToolServer) (*ToolServer, error)
	RefreshToolsForServer(serverName string, tools []*api.Component) error
	DeleteRun(runID int) error
	DeleteSession(sessionName string, userID string) error
	DeleteTeam(teamName string) error
	DeleteToolServer(serverName string) error
	GetRun(runID int) (*Run, error)
	GetRunMessages(runID int) ([]*Message, error)
	GetSession(sessionLabel string, userID string) (*Session, error)
	GetTeam(teamLabel string) (*Team, error)
	GetTool(provider string) (*Tool, error)
	GetToolServer(serverName string) (*ToolServer, error)
	ListFeedback(userID string) ([]*Feedback, error)
	ListRuns(userID string) ([]*Run, error)
	ListSessionRuns(sessionName string, userID string) ([]*Run, error)
	ListSessions(userID string) ([]*Session, error)
	ListTeams(userID string) ([]*Team, error)
	ListToolServers() ([]*ToolServer, error)
	ListTools(userID string) ([]*Tool, error)
	ListToolsForServer(serverName string) ([]*Tool, error)
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
func (c *clientImpl) GetRunMessages(runID int) ([]*Message, error) {
	messages, err := c.serviceWrapper.Message.List(Clause{Key: "run_id", Value: runID})
	if err != nil {
		return nil, err
	}

	// Convert []Message to []*Message
	messagePtrs := make([]*Message, len(messages))
	for i := range messages {
		messagePtrs[i] = &messages[i]
	}
	return messagePtrs, nil
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
func (c *clientImpl) ListFeedback(userID string) ([]*Feedback, error) {
	feedback, err := c.serviceWrapper.Feedback.List(Clause{Key: "user_id", Value: userID})
	if err != nil {
		return nil, err
	}

	// Convert []Feedback to []*Feedback
	feedbackPtrs := make([]*Feedback, len(feedback))
	for i := range feedback {
		feedbackPtrs[i] = &feedback[i]
	}
	return feedbackPtrs, nil
}

// ListRuns lists all runs for a user
func (c *clientImpl) ListRuns(userID string) ([]*Run, error) {
	runs, err := c.serviceWrapper.Run.List(Clause{Key: "user_id", Value: userID})
	if err != nil {
		return nil, err
	}

	// Convert []Run to []*Run
	runPtrs := make([]*Run, len(runs))
	for i := range runs {
		runPtrs[i] = &runs[i]
	}
	return runPtrs, nil
}

// ListSessionRuns lists all runs for a specific session
func (c *clientImpl) ListSessionRuns(sessionName string, userID string) ([]*Run, error) {
	runs, err := c.serviceWrapper.Run.List(
		Clause{Key: "session_name", Value: sessionName},
		Clause{Key: "user_id", Value: userID},
	)
	if err != nil {
		return nil, err
	}

	// Convert []Run to []*Run
	runPtrs := make([]*Run, len(runs))
	for i := range runs {
		runPtrs[i] = &runs[i]
	}
	return runPtrs, nil
}

// ListSessions lists all sessions for a user
func (c *clientImpl) ListSessions(userID string) ([]*Session, error) {
	sessions, err := c.serviceWrapper.Session.List(Clause{Key: "user_id", Value: userID})
	if err != nil {
		return nil, err
	}

	// Convert []Session to []*Session
	sessionPtrs := make([]*Session, len(sessions))
	for i := range sessions {
		sessionPtrs[i] = &sessions[i]
	}
	return sessionPtrs, nil
}

// ListTeams lists all teams for a user
func (c *clientImpl) ListTeams(userID string) ([]*Team, error) {
	teams, err := c.serviceWrapper.Team.List()
	if err != nil {
		return nil, err
	}

	// Convert []Team to []*Team
	teamPtrs := make([]*Team, len(teams))
	for i := range teams {
		teamPtrs[i] = &teams[i]
	}
	return teamPtrs, nil
}

// ListToolServers lists all tool servers for a user
func (c *clientImpl) ListToolServers() ([]*ToolServer, error) {
	servers, err := c.serviceWrapper.ToolServer.List()
	if err != nil {
		return nil, err
	}

	// Convert []ToolServer to []*ToolServer
	serverPtrs := make([]*ToolServer, len(servers))
	for i := range servers {
		serverPtrs[i] = &servers[i]
	}
	return serverPtrs, nil
}

// ListTools lists all tools for a user
func (c *clientImpl) ListTools(userID string) ([]*Tool, error) {
	tools, err := c.serviceWrapper.Tool.List()
	if err != nil {
		return nil, err
	}

	// Convert []Tool to []*Tool
	toolPtrs := make([]*Tool, len(tools))
	for i := range tools {
		toolPtrs[i] = &tools[i]
	}
	return toolPtrs, nil
}

// ListToolsForServer lists all tools for a specific server
func (c *clientImpl) ListToolsForServer(serverName string) ([]*Tool, error) {
	tools, err := c.serviceWrapper.Tool.List(Clause{Key: "server_name", Value: serverName})
	if err != nil {
		return nil, err
	}

	// Convert []Tool to []*Tool
	toolPtrs := make([]*Tool, len(tools))
	for i := range tools {
		toolPtrs[i] = &tools[i]
	}
	return toolPtrs, nil
}

// RefreshToolsForServer refreshes a tool server (placeholder implementation)
func (c *clientImpl) RefreshToolsForServer(serverName string, tools []*api.Component) error {
	existingTools, err := c.ListToolsForServer(serverName)
	if err != nil {
		return err
	}

	// Check if the tool exists in the existing tools
	// If it does, update it
	// If it doesn't, create it
	// If it's in the existing tools but not in the new tools, delete it
	for _, tool := range tools {
		existingToolIndex := slices.IndexFunc(existingTools, func(t *Tool) bool {
			return t.Name == tool.Label
		})
		if existingToolIndex != -1 {
			existingTool := existingTools[existingToolIndex]
			existingTool.Component = *tool
			err = c.serviceWrapper.Tool.Update(existingTool)
			if err != nil {
				return err
			}
		} else {
			err = c.serviceWrapper.Tool.Create(&Tool{
				Name:      tool.Label,
				Component: *tool,
			})
			if err != nil {
				return fmt.Errorf("failed to create tool %s: %v", tool.Label, err)
			}
		}
	}

	// Delete any tools that are in the existing tools but not in the new tools
	for _, existingTool := range existingTools {
		if !slices.ContainsFunc(tools, func(t *api.Component) bool {
			return t.Label == existingTool.Name
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
