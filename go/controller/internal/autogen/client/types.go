package client

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/kagent-dev/kagent/go/controller/internal/autogen/api"
)

type ModelsUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

func (m *ModelsUsage) Add(other *ModelsUsage) {
	if other == nil {
		return
	}
	m.PromptTokens += other.PromptTokens
	m.CompletionTokens += other.CompletionTokens
}

func (m *ModelsUsage) String() string {
	return fmt.Sprintf("Prompt Tokens: %d, Completion Tokens: %d", m.PromptTokens, m.CompletionTokens)
}

type TaskMessageMap map[string]interface{}

type RunMessage struct {
	CreatedAt   *string                `json:"created_at,omitempty"`
	UpdatedAt   *string                `json:"updated_at,omitempty"`
	Version     *string                `json:"version,omitempty"`
	SessionID   int                    `json:"session_id"`
	MessageMeta map[string]interface{} `json:"message_meta"`
	ID          int                    `json:"id"`
	UserID      *string                `json:"user_id"`
	Config      map[string]interface{} `json:"config"`
	RunID       int                    `json:"run_id"`
}

type CreateRunRequest struct {
	SessionID int    `json:"session_id"`
	UserID    string `json:"user_id"`
}

type CreateRunResult struct {
	ID int `json:"run_id"`
}

type SessionRuns struct {
	Runs []Run `json:"runs"`
}

type Run struct {
	ID           int           `json:"id"`
	SessionID    int           `json:"session_id"`
	CreatedAt    string        `json:"created_at"`
	Status       string        `json:"status"`
	Task         Task          `json:"task"`
	TeamResult   TeamResult    `json:"team_result"`
	Messages     []*RunMessage `json:"messages"`
	ErrorMessage string        `json:"error_message"`
}

type Task struct {
	Source      string      `json:"source"`
	Content     interface{} `json:"content"`
	MessageType string      `json:"message_type"`
}

type TeamResult struct {
	TaskResult TaskResult `json:"task_result"`
	Usage      string     `json:"usage"`
	Duration   float64    `json:"duration"`
}

type TaskResult struct {
	Messages   []TaskMessageMap `json:"messages"`
	StopReason string           `json:"stop_reason"`
}

// APIResponse is the common response wrapper for all API responses
type APIResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Session struct {
	ID        int    `json:"id"`
	UserID    string `json:"user_id"`
	Version   string `json:"version"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	TeamID    *int   `json:"team_id"`
}

type CreateSession struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	TeamID *int   `json:"team_id"`
}

// ProviderModels maps provider names to a list of their supported model names.
type ProviderModels map[string][]ModelInfo

// ModelInfo holds details about a specific model.
type ModelInfo struct {
	Name            string `json:"name"`
	FunctionCalling bool   `json:"function_calling"`
}

type SseEvent struct {
	Event string `json:"event"`
	Data  []byte `json:"data"`
}

// InvokeRequest represents the request payload for session invocation
type InvokeRequest struct {
	Task       string         `json:"task"`
	TeamConfig *api.Component `json:"team_config"`
}

var (
	NotFoundError = errors.New("not found")
)

func streamSseResponse(r io.ReadCloser) chan *SseEvent {
	scanner := bufio.NewScanner(r)
	ch := make(chan *SseEvent)
	go func() {
		defer close(ch)
		defer r.Close()
		currentEvent := &SseEvent{}
		for scanner.Scan() {
			line := scanner.Bytes()
			if bytes.HasPrefix(line, []byte("event:")) {
				currentEvent.Event = string(bytes.TrimPrefix(line, []byte("event:")))
			}
			if bytes.HasPrefix(line, []byte("data:")) {
				currentEvent.Data = bytes.TrimPrefix(line, []byte("data:"))
				ch <- currentEvent
				currentEvent = &SseEvent{}
			}
		}
	}()
	return ch
}
