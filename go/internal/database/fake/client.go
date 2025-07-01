package fake

import (
	"fmt"
	"sync"

	"github.com/kagent-dev/kagent/go/internal/database"
)

// Client is a fake implementation of database.Client for testing
type Client struct {
	mu             sync.RWMutex
	feedback       map[string]*database.Feedback
	runs           map[int]*database.Run
	sessions       map[string]*database.Session // key: sessionName_userID
	teams          map[string]*database.Team
	toolServers    map[string]*database.ToolServer
	tools          map[string]*database.Tool
	messages       map[int][]*database.Message // key: runID
	nextRunID      int
	nextFeedbackID int
}

// NewClient creates a new fake database client
func NewClient() database.Client {
	return &Client{
		feedback:       make(map[string]*database.Feedback),
		runs:           make(map[int]*database.Run),
		sessions:       make(map[string]*database.Session),
		teams:          make(map[string]*database.Team),
		toolServers:    make(map[string]*database.ToolServer),
		tools:          make(map[string]*database.Tool),
		messages:       make(map[int][]*database.Message),
		nextRunID:      1,
		nextFeedbackID: 1,
	}
}

func (c *Client) sessionKey(name, userID string) string {
	return fmt.Sprintf("%s_%s", name, userID)
}

// CreateFeedback creates a new feedback record
func (c *Client) CreateFeedback(feedback *database.Feedback) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Copy the feedback and assign an ID
	newFeedback := *feedback
	newFeedback.ID = uint(c.nextFeedbackID)
	c.nextFeedbackID++

	key := fmt.Sprintf("%d", newFeedback.ID)
	c.feedback[key] = &newFeedback
	return nil
}

// CreateRun creates a new run record
func (c *Client) CreateRun(req *database.Run) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Copy the run and assign an ID
	newRun := *req
	newRun.ID = uint(c.nextRunID)
	c.nextRunID++

	c.runs[int(newRun.ID)] = &newRun
	return nil
}

// CreateSession creates a new session record
func (c *Client) CreateSession(session *database.Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.sessionKey(session.Name, session.UserID)
	c.sessions[key] = session
	return nil
}

// CreateTeam creates a new team record
func (c *Client) CreateTeam(team *database.Team) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.teams[team.Name] = team
	return nil
}

// UpsertTeam upserts a team record
func (c *Client) UpsertTeam(team *database.Team) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.teams[team.Name] = team
	return nil
}

// CreateToolServer creates a new tool server record
func (c *Client) CreateToolServer(toolServer *database.ToolServer) (*database.ToolServer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.toolServers[toolServer.Name] = toolServer
	return toolServer, nil
}

// DeleteRun deletes a run by ID
func (c *Client) DeleteRun(runID int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.runs, runID)
	delete(c.messages, runID)
	return nil
}

// DeleteSession deletes a session by name and user ID
func (c *Client) DeleteSession(sessionName string, userID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.sessionKey(sessionName, userID)
	delete(c.sessions, key)
	return nil
}

// DeleteTeam deletes a team by name
func (c *Client) DeleteTeam(teamName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.teams, teamName)
	return nil
}

// DeleteToolServer deletes a tool server by name
func (c *Client) DeleteToolServer(serverName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.toolServers, serverName)
	return nil
}

// GetRun retrieves a run by ID
func (c *Client) GetRun(runID int) (*database.Run, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	run, exists := c.runs[runID]
	if !exists {
		return nil, fmt.Errorf("run with ID %d not found", runID)
	}
	return run, nil
}

// GetRunMessages retrieves messages for a specific run
func (c *Client) GetRunMessages(runID int) ([]*database.Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages, exists := c.messages[runID]
	if !exists {
		return []*database.Message{}, nil
	}
	return messages, nil
}

// GetSession retrieves a session by name and user ID
func (c *Client) GetSession(sessionLabel string, userID string) (*database.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.sessionKey(sessionLabel, userID)
	session, exists := c.sessions[key]
	if !exists {
		return nil, fmt.Errorf("session with label %s for user %s not found", sessionLabel, userID)
	}
	return session, nil
}

// GetTeam retrieves a team by name
func (c *Client) GetTeam(teamLabel string) (*database.Team, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	team, exists := c.teams[teamLabel]
	if !exists {
		return nil, fmt.Errorf("team with label %s not found", teamLabel)
	}
	return team, nil
}

// GetTool retrieves a tool by provider
func (c *Client) GetTool(provider string) (*database.Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tool, exists := c.tools[provider]
	if !exists {
		return nil, fmt.Errorf("tool with provider %s not found", provider)
	}
	return tool, nil
}

// GetToolServer retrieves a tool server by name
func (c *Client) GetToolServer(serverName string) (*database.ToolServer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	server, exists := c.toolServers[serverName]
	if !exists {
		return nil, fmt.Errorf("tool server with name %s not found", serverName)
	}
	return server, nil
}

// ListFeedback lists all feedback for a user
func (c *Client) ListFeedback(userID string) ([]*database.Feedback, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Feedback
	for _, feedback := range c.feedback {
		if feedback.UserID == userID {
			result = append(result, feedback)
		}
	}
	return result, nil
}

// ListRuns lists all runs for a user
func (c *Client) ListRuns(userID string) ([]*database.Run, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Run
	for _, run := range c.runs {
		if run.UserID == userID {
			result = append(result, run)
		}
	}
	return result, nil
}

// ListSessionRuns lists all runs for a specific session
func (c *Client) ListSessionRuns(sessionName string, userID string) ([]*database.Run, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Run
	for _, run := range c.runs {
		if run.SessionName == sessionName && run.UserID == userID {
			result = append(result, run)
		}
	}
	return result, nil
}

// ListSessions lists all sessions for a user
func (c *Client) ListSessions(userID string) ([]*database.Session, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Session
	for _, session := range c.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}
	return result, nil
}

// ListTeams lists all teams for a user
func (c *Client) ListTeams(userID string) ([]*database.Team, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Team
	for _, team := range c.teams {
		result = append(result, team)
	}
	return result, nil
}

// ListToolServers lists all tool servers
func (c *Client) ListToolServers() ([]*database.ToolServer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.ToolServer
	for _, server := range c.toolServers {
		result = append(result, server)
	}
	return result, nil
}

// ListTools lists all tools for a user
func (c *Client) ListTools(userID string) ([]*database.Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Tool
	for _, tool := range c.tools {
		result = append(result, tool)
	}
	return result, nil
}

// ListToolsForServer lists all tools for a specific server
func (c *Client) ListToolsForServer(serverName string) ([]*database.Tool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*database.Tool
	for _, tool := range c.tools {
		if tool.ServerName == serverName {
			result = append(result, tool)
		}
	}
	return result, nil
}

// ListMessagesForRun retrieves messages for a specific run
func (c *Client) ListMessagesForRun(runID uint) ([]database.Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages, exists := c.messages[int(runID)]
	if !exists {
		return []database.Message{}, nil
	}

	// Convert []*Message to []Message
	result := make([]database.Message, len(messages))
	for i, msg := range messages {
		result[i] = *msg
	}
	return result, nil
}

// UpdateSession updates a session
func (c *Client) UpdateSession(session *database.Session) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.sessionKey(session.Name, session.UserID)
	c.sessions[key] = session
	return nil
}

// UpdateToolServer updates a tool server
func (c *Client) UpdateToolServer(server *database.ToolServer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.toolServers[server.Name] = server
	return nil
}

// UpdateRun updates a run record
func (c *Client) UpdateRun(run *database.Run) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.runs[int(run.ID)] = run
	return nil
}

// UpdateTeam updates a team record
func (c *Client) UpdateTeam(team *database.Team) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.teams[team.Name] = team
	return nil
}

// Helper methods for testing

// AddMessage adds a message to a run for testing purposes
func (c *Client) AddMessage(runID int, message *database.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.messages[runID] == nil {
		c.messages[runID] = []*database.Message{}
	}
	c.messages[runID] = append(c.messages[runID], message)
}

// AddTool adds a tool for testing purposes
func (c *Client) AddTool(tool *database.Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tools[tool.Name] = tool
}

// Clear clears all data for testing purposes
func (c *Client) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.feedback = make(map[string]*database.Feedback)
	c.runs = make(map[int]*database.Run)
	c.sessions = make(map[string]*database.Session)
	c.teams = make(map[string]*database.Team)
	c.toolServers = make(map[string]*database.ToolServer)
	c.tools = make(map[string]*database.Tool)
	c.messages = make(map[int][]*database.Message)
	c.nextRunID = 1
	c.nextFeedbackID = 1
}
