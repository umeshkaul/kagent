package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Service provides high-level database operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new database service
func NewService(manager *Manager) *Service {
	return &Service{db: manager.db}
}

// Teams operations

// ListTeams retrieves all teams for a user
func (s *Service) ListTeams(userID string) ([]Team, error) {
	var teams []Team
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&teams).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	return teams, nil
}

// GetTeam retrieves a specific team by ID and user
func (s *Service) GetTeam(teamID uint, userID string) (*Team, error) {
	var team Team
	err := s.db.Where("id = ? AND user_id = ?", teamID, userID).First(&team).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return &team, nil
}

// CreateTeam creates a new team
func (s *Service) CreateTeam(team *Team) error {
	err := s.db.Create(team).Error
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}
	return nil
}

// UpdateTeam updates an existing team
func (s *Service) UpdateTeam(team *Team) error {
	err := s.db.Save(team).Error
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}
	return nil
}

// DeleteTeam deletes a team
func (s *Service) DeleteTeam(teamID uint, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", teamID, userID).Delete(&Team{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete team: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}

// Sessions operations

// ListSessions retrieves all sessions for a user
func (s *Service) ListSessions(userID string) ([]Session, error) {
	var sessions []Session
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, nil
}

// GetSession retrieves a specific session by ID and user
func (s *Service) GetSession(sessionID uint, userID string) (*Session, error) {
	var session Session
	err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// CreateSession creates a new session
func (s *Service) CreateSession(session *Session) error {
	err := s.db.Create(session).Error
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// UpdateSession updates an existing session
func (s *Service) UpdateSession(session *Session) error {
	err := s.db.Save(session).Error
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	return nil
}

// DeleteSession deletes a session
func (s *Service) DeleteSession(sessionID uint, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&Session{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// GetSessionRuns retrieves runs for a specific session
func (s *Service) GetSessionRuns(sessionID uint, userID string) ([]Run, error) {
	// First verify the session belongs to the user
	var session Session
	err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found or access denied")
		}
		return nil, fmt.Errorf("failed to verify session: %w", err)
	}

	// Get runs for the session with preloaded messages
	var runs []Run
	err = s.db.Preload("MessageItems").Where("session_id = ?", sessionID).Order("created_at ASC").Find(&runs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get session runs: %w", err)
	}

	return runs, nil
}

// Runs operations

// CreateRun creates a new run
func (s *Service) CreateRun(run *Run) error {
	err := s.db.Create(run).Error
	if err != nil {
		return fmt.Errorf("failed to create run: %w", err)
	}
	return nil
}

// UpdateRun updates an existing run
func (s *Service) UpdateRun(run *Run) error {
	err := s.db.Save(run).Error
	if err != nil {
		return fmt.Errorf("failed to update run: %w", err)
	}
	return nil
}

// GetRun retrieves a specific run by ID
func (s *Service) GetRun(runID uint, userID string) (*Run, error) {
	var run Run
	err := s.db.Where("id = ? AND user_id = ?", runID, userID).First(&run).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("run not found")
		}
		return nil, fmt.Errorf("failed to get run: %w", err)
	}
	return &run, nil
}

// Messages operations

// CreateMessage creates a new message
func (s *Service) CreateMessage(message *Message) error {
	err := s.db.Create(message).Error
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	return nil
}

// GetMessagesForRun retrieves messages for a specific run
func (s *Service) GetMessagesForRun(runID uint) ([]Message, error) {
	var messages []Message
	err := s.db.Where("run_id = ?", runID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}

// Tools operations

// ListTools retrieves all tools for a user
func (s *Service) ListTools(userID string) ([]Tool, error) {
	var tools []Tool
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tools).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}
	return tools, nil
}

// CreateTool creates a new tool
func (s *Service) CreateTool(tool *Tool) error {
	err := s.db.Create(tool).Error
	if err != nil {
		return fmt.Errorf("failed to create tool: %w", err)
	}
	return nil
}

// UpdateTool updates an existing tool
func (s *Service) UpdateTool(tool *Tool) error {
	err := s.db.Save(tool).Error
	if err != nil {
		return fmt.Errorf("failed to update tool: %w", err)
	}
	return nil
}

// DeleteTool deletes a tool
func (s *Service) DeleteTool(toolID uint, userID string) error {
	result := s.db.Where("id = ? AND user_id = ?", toolID, userID).Delete(&Tool{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete tool: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tool not found")
	}
	return nil
}

// GetTool retrieves a specific tool by ID and user
func (s *Service) GetTool(toolID uint, userID string) (*Tool, error) {
	var tool Tool
	err := s.db.Where("id = ? AND user_id = ?", toolID, userID).First(&tool).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tool not found")
		}
		return nil, fmt.Errorf("failed to get tool: %w", err)
	}
	return &tool, nil
}

// ToolServers operations

// ListToolServers retrieves all tool servers for a user
func (s *Service) ListToolServers(userID string) ([]ToolServer, error) {
	var servers []ToolServer
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&servers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list tool servers: %w", err)
	}
	return servers, nil
}

// GetToolServer retrieves a specific tool server by ID and user
func (s *Service) GetToolServer(serverID uint, userID string) (*ToolServer, error) {
	var server ToolServer
	err := s.db.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tool server not found")
		}
		return nil, fmt.Errorf("failed to get tool server: %w", err)
	}
	return &server, nil
}

// CreateToolServer creates a new tool server
func (s *Service) CreateToolServer(server *ToolServer) error {
	err := s.db.Create(server).Error
	if err != nil {
		return fmt.Errorf("failed to create tool server: %w", err)
	}
	return nil
}

// UpdateToolServer updates an existing tool server
func (s *Service) UpdateToolServer(server *ToolServer) error {
	now := time.Now()
	server.LastConnected = &now
	err := s.db.Save(server).Error
	if err != nil {
		return fmt.Errorf("failed to update tool server: %w", err)
	}
	return nil
}

// DeleteToolServer deletes a tool server and its associated tools
func (s *Service) DeleteToolServer(serverID uint, userID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// First delete associated tools
		result := tx.Where("server_id = ? AND user_id = ?", serverID, userID).Delete(&Tool{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete associated tools: %w", result.Error)
		}

		// Then delete the server
		result = tx.Where("id = ? AND user_id = ?", serverID, userID).Delete(&ToolServer{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete tool server: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("tool server not found")
		}
		return nil
	})
}

// GetToolsForServer retrieves tools for a specific server
func (s *Service) GetToolsForServer(serverID uint, userID string) ([]Tool, error) {
	var tools []Tool
	err := s.db.Where("server_id = ? AND user_id = ?", serverID, userID).Order("created_at DESC").Find(&tools).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tools for server: %w", err)
	}
	return tools, nil
}

// Feedback operations

// CreateFeedback creates a new feedback entry
func (s *Service) CreateFeedback(feedback *Feedback) error {
	err := s.db.Create(feedback).Error
	if err != nil {
		return fmt.Errorf("failed to create feedback: %w", err)
	}
	return nil
}

// ListFeedback retrieves all feedback for a user
func (s *Service) ListFeedback(userID string) ([]Feedback, error) {
	var feedback []Feedback
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&feedback).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback: %w", err)
	}
	return feedback, nil
}

// Gallery operations

// GetGallery retrieves gallery configuration for a user
func (s *Service) GetGallery(userID string) (*Gallery, error) {
	var gallery Gallery
	err := s.db.Where("user_id = ?", userID).First(&gallery).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("gallery not found")
		}
		return nil, fmt.Errorf("failed to get gallery: %w", err)
	}
	return &gallery, nil
}

// CreateGallery creates a new gallery configuration
func (s *Service) CreateGallery(gallery *Gallery) error {
	err := s.db.Create(gallery).Error
	if err != nil {
		return fmt.Errorf("failed to create gallery: %w", err)
	}
	return nil
}

// Settings operations

// GetSettings retrieves settings for a user
func (s *Service) GetSettings(userID string) (*Settings, error) {
	var settings Settings
	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("settings not found")
		}
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	return &settings, nil
}

// CreateSettings creates new settings
func (s *Service) CreateSettings(settings *Settings) error {
	err := s.db.Create(settings).Error
	if err != nil {
		return fmt.Errorf("failed to create settings: %w", err)
	}
	return nil
}

// UpdateSettings updates existing settings
func (s *Service) UpdateSettings(settings *Settings) error {
	err := s.db.Save(settings).Error
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	return nil
}
