package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/kagent-dev/kagent/go/internal/autogen/api"
	"gorm.io/gorm"
)

// JSONMap is a custom type for handling JSON columns in GORM
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONMap: value is not []byte")
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Team represents a team configuration
type Team struct {
	gorm.Model
	Name      string        `gorm:"unique;not null" json:"name"`
	Component api.Component `gorm:"type:json;not null" json:"component"`
}

// Session represents a conversation session
type Session struct {
	gorm.Model
	UserID string `gorm:"primaryKey;not null" json:"user_id"`
	Name   string `gorm:"unique;not null" json:"name"`

	// Relationships
	Runs []Run `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"runs,omitempty"`

	TeamID *uint `gorm:"index;constraint:OnDelete:SET NULL" json:"team_id,omitempty"`
}

// RunStatus represents the status of a run
type RunStatus string

const (
	RunStatusCreated  RunStatus = "created"
	RunStatusActive   RunStatus = "active"
	RunStatusComplete RunStatus = "complete"
	RunStatusError    RunStatus = "error"
	RunStatusStopped  RunStatus = "stopped"
)

// Run represents a single execution run within a session
type Run struct {
	gorm.Model
	UserID       string    `gorm:"primaryKey;not null" json:"user_id"`
	SessionID    uint      `gorm:"not null;index;constraint:OnDelete:CASCADE" json:"session_id"`
	Status       RunStatus `gorm:"default:created" json:"status"`
	Task         JSONMap   `gorm:"type:json;not null" json:"task"`
	TeamResult   JSONMap   `gorm:"type:json" json:"team_result,omitempty"`
	ErrorMessage *string   `json:"error_message,omitempty"`

	// Relationships
	MessageItems []Message `gorm:"foreignKey:RunID;constraint:OnDelete:CASCADE" json:"message_items,omitempty"`
}

// Message represents a message in a conversation
type Message struct {
	gorm.Model
	UserID      string  `gorm:"primaryKey;not null" json:"user_id"`
	Config      JSONMap `gorm:"type:json;not null" json:"config"`
	SessionID   uint    `gorm:"index;constraint:OnDelete:SET NULL" json:"session_id,omitempty"`
	RunID       uint    `gorm:"index;constraint:OnDelete:CASCADE" json:"run_id,omitempty"`
	MessageMeta JSONMap `gorm:"type:json" json:"message_meta,omitempty"`

	// Relationships
	Feedback []Feedback `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE" json:"feedback,omitempty"`
}

// FeedbackIssueType represents the category of feedback issue
type FeedbackIssueType string

const (
	FeedbackIssueTypeInstructions FeedbackIssueType = "instructions" // Did not follow instructions
	FeedbackIssueTypeFactual      FeedbackIssueType = "factual"      // Not factually correct
	FeedbackIssueTypeIncomplete   FeedbackIssueType = "incomplete"   // Incomplete response
	FeedbackIssueTypeTool         FeedbackIssueType = "tool"         // Should have run the tool
)

// Feedback represents user feedback on agent responses
type Feedback struct {
	gorm.Model
	UserID       string             `gorm:"primaryKey;not null" json:"user_id"`
	MessageID    uint               `gorm:"index;constraint:OnDelete:CASCADE" json:"message_id"`
	IsPositive   bool               `gorm:"default:false" json:"is_positive"`
	FeedbackText string             `gorm:"not null" json:"feedback_text"`
	IssueType    *FeedbackIssueType `json:"issue_type,omitempty"`
}

// Tool represents a single tool that can be used by an agent
type Tool struct {
	gorm.Model
	Name      string        `gorm:"unique" json:"name"`
	Component api.Component `gorm:"type:json;not null" json:"component"`
	ServerID  uint          `gorm:"index;constraint:OnDelete:SET NULL" json:"server_id,omitempty"`

	// Relationships
	ToolServer *ToolServer `gorm:"foreignKey:ServerName" json:"tool_server,omitempty"`
}

// ToolServer represents a tool server that provides tools
type ToolServer struct {
	gorm.Model
	Name          string        `gorm:"primaryKey" json:"name"`
	LastConnected *time.Time    `json:"last_connected,omitempty"`
	Component     api.Component `gorm:"type:json;not null" json:"component"`
}

// EvalTask represents an evaluation task
type EvalTask struct {
	gorm.Model
	UserID      string        `gorm:"primaryKey;not null" json:"user_id"`
	Name        string        `gorm:"default:'Unnamed Task'" json:"name"`
	Description string        `json:"description"`
	Config      api.Component `gorm:"type:json;not null" json:"config"`
}

// EvalCriteria represents evaluation criteria
type EvalCriteria struct {
	gorm.Model
	Name        string        `gorm:"default:'Unnamed Criteria'" json:"name"`
	Description string        `json:"description"`
	Config      api.Component `gorm:"type:json;not null" json:"config"`
}

// EvalRunStatus represents the status of an evaluation run
type EvalRunStatus string

const (
	EvalRunStatusPending  EvalRunStatus = "pending"
	EvalRunStatusRunning  EvalRunStatus = "running"
	EvalRunStatusComplete EvalRunStatus = "complete"
	EvalRunStatusError    EvalRunStatus = "error"
)

// EvalRun represents an evaluation run
type EvalRun struct {
	gorm.Model
	UserID          string          `gorm:"primaryKey;not null" json:"user_id"`
	Name            string          `gorm:"default:'Unnamed Evaluation Run'" json:"name"`
	Description     string          `json:"description"`
	TaskID          *uint           `gorm:"index;constraint:OnDelete:SET NULL" json:"task_id,omitempty"`
	RunnerConfig    api.Component   `gorm:"not null" json:"runner_config"`
	JudgeConfig     api.Component   `gorm:"not null" json:"judge_config"`
	CriteriaConfigs []api.Component `json:"criteria_configs"`
	Status          EvalRunStatus   `gorm:"default:pending" json:"status"`
	StartTime       *time.Time      `json:"start_time,omitempty"`
	EndTime         *time.Time      `json:"end_time,omitempty"`
	RunResult       JSONMap         `gorm:"type:json" json:"run_result,omitempty"`
	ScoreResult     JSONMap         `gorm:"type:json" json:"score_result,omitempty"`
	ErrorMessage    *string         `json:"error_message,omitempty"`

	// Relationships
	Task *EvalTask `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// TableName methods to match Python table names
func (Team) TableName() string         { return "team" }
func (Session) TableName() string      { return "session" }
func (Run) TableName() string          { return "run" }
func (Message) TableName() string      { return "message" }
func (Feedback) TableName() string     { return "feedback" }
func (Tool) TableName() string         { return "tool" }
func (ToolServer) TableName() string   { return "toolserver" }
func (EvalTask) TableName() string     { return "evaltask" }
func (EvalCriteria) TableName() string { return "evalcriteria" }
func (EvalRun) TableName() string      { return "evalrun" }
