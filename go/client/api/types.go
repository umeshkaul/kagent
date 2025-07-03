package api

import (
	"time"

	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
	"github.com/kagent-dev/kagent/go/internal/autogen/api"
	"github.com/kagent-dev/kagent/go/internal/database"
)

// Common types

// APIError represents an error response from the API
type APIError struct {
	Error string `json:"error"`
}

func NewResponse[T any](data T, message string, error bool) StandardResponse[T] {
	return StandardResponse[T]{
		Error:   error,
		Data:    data,
		Message: message,
	}
}

// StandardResponse represents the standard response format used by many endpoints
type StandardResponse[T any] struct {
	Error   bool   `json:"error"`
	Data    T      `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// Provider represents a provider configuration
type Provider struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Version represents the version information
type VersionResponse struct {
	KAgentVersion string `json:"kagent_version"`
	GitCommit     string `json:"git_commit"`
	BuildDate     string `json:"build_date"`
}

// ModelConfigResponse represents a model configuration response
type ModelConfigResponse struct {
	Ref             string                 `json:"ref"`
	ProviderName    string                 `json:"providerName"`
	Model           string                 `json:"model"`
	APIKeySecretRef string                 `json:"apiKeySecretRef"`
	APIKeySecretKey string                 `json:"apiKeySecretKey"`
	ModelParams     map[string]interface{} `json:"modelParams"`
}

// CreateModelConfigRequest represents a request to create a model configuration
type CreateModelConfigRequest struct {
	Ref             string                      `json:"ref"`
	Provider        Provider                    `json:"provider"`
	Model           string                      `json:"model"`
	APIKey          string                      `json:"apiKey"`
	OpenAIParams    *v1alpha1.OpenAIConfig      `json:"openAI,omitempty"`
	AnthropicParams *v1alpha1.AnthropicConfig   `json:"anthropic,omitempty"`
	AzureParams     *v1alpha1.AzureOpenAIConfig `json:"azureOpenAI,omitempty"`
	OllamaParams    *v1alpha1.OllamaConfig      `json:"ollama,omitempty"`
}

// UpdateModelConfigRequest represents a request to update a model configuration
type UpdateModelConfigRequest struct {
	Provider        Provider                    `json:"provider"`
	Model           string                      `json:"model"`
	APIKey          *string                     `json:"apiKey,omitempty"`
	OpenAIParams    *v1alpha1.OpenAIConfig      `json:"openAI,omitempty"`
	AnthropicParams *v1alpha1.AnthropicConfig   `json:"anthropic,omitempty"`
	AzureParams     *v1alpha1.AzureOpenAIConfig `json:"azureOpenAI,omitempty"`
	OllamaParams    *v1alpha1.OllamaConfig      `json:"ollama,omitempty"`
}

// Session types

// SessionRequest represents a session creation/update request
type SessionRequest struct {
	TeamID *uint  `json:"team_id,omitempty"`
	Name   string `json:"name"`
	UserID string `json:"user_id"`
}

// Session represents a session from the database
type Session struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	TeamID    *uint     `json:"team_id,omitempty"`
	Runs      []Run     `json:"runs,omitempty"`
}

// Run types

// RunRequest represents a run creation request
type RunRequest struct {
	Task string `json:"task"`
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

// Run represents a run from the database
type Run = database.Run

// Message represents a message from the database
type Message = database.Message

// Team types

// TeamRequest represents a team creation/update request
type TeamRequest struct {
	AgentRef  string        `json:"agent_ref"`
	Component api.Component `json:"component"`
}

// Team represents a team from the database
type Team = database.Team

// Tool types

// Tool represents a tool from the database
type Tool = database.Tool

// ToolServer types

// ToolServerResponse represents a tool server response
type ToolServerResponse struct {
	Ref             string                    `json:"ref"`
	Config          v1alpha1.ToolServerConfig `json:"config"`
	DiscoveredTools []*v1alpha1.MCPTool       `json:"discoveredTools"`
}

// Memory types

// MemoryResponse represents a memory response
type MemoryResponse struct {
	Ref             string                 `json:"ref"`
	ProviderName    string                 `json:"providerName"`
	APIKeySecretRef string                 `json:"apiKeySecretRef"`
	APIKeySecretKey string                 `json:"apiKeySecretKey"`
	MemoryParams    map[string]interface{} `json:"memoryParams"`
}

// CreateMemoryRequest represents a request to create a memory
type CreateMemoryRequest struct {
	Ref            string                   `json:"ref"`
	Provider       Provider                 `json:"provider"`
	APIKey         string                   `json:"apiKey"`
	PineconeParams *v1alpha1.PineconeConfig `json:"pinecone,omitempty"`
}

// UpdateMemoryRequest represents a request to update a memory
type UpdateMemoryRequest struct {
	PineconeParams *v1alpha1.PineconeConfig `json:"pinecone,omitempty"`
}

// Namespace types

// NamespaceResponse represents a namespace response
type NamespaceResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Feedback types

// FeedbackIssueType represents the category of feedback issue
type FeedbackIssueType string

const (
	FeedbackIssueTypeInstructions FeedbackIssueType = "instructions" // Did not follow instructions
	FeedbackIssueTypeFactual      FeedbackIssueType = "factual"      // Not factually correct
	FeedbackIssueTypeIncomplete   FeedbackIssueType = "incomplete"   // Incomplete response
	FeedbackIssueTypeTool         FeedbackIssueType = "tool"         // Should have run the tool
)

// Feedback represents feedback from the database
type Feedback struct {
	ID           uint               `json:"id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	UserID       string             `json:"user_id"`
	MessageID    uint               `json:"message_id"`
	IsPositive   bool               `json:"is_positive"`
	FeedbackText string             `json:"feedback_text"`
	IssueType    *FeedbackIssueType `json:"issue_type,omitempty"`
}

// Provider types

// ProviderInfo represents information about a provider
type ProviderInfo struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	RequiredParams []string `json:"requiredParams"`
	OptionalParams []string `json:"optionalParams"`
}

// SessionRunsResponse represents the response for session runs
type SessionRunsResponse struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

// SessionRunsData represents the data part of session runs response
type SessionRunsData struct {
	Runs []interface{} `json:"runs"`
}
