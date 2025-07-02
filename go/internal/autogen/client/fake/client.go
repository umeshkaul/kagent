package fake

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	autogen_client "github.com/kagent-dev/kagent/go/internal/autogen/client"
)

type InMemoryAutogenClient struct {
	mu sync.RWMutex

	// Minimal storage for FetchTools functionality
	toolsByServer map[string][]*autogen_client.NamedTool
}

func NewInMemoryAutogenClient() *InMemoryAutogenClient {
	return &InMemoryAutogenClient{
		toolsByServer: make(map[string][]*autogen_client.NamedTool),
	}
}

// NewMockAutogenClient creates a new in-memory autogen client for backward compatibility
func NewMockAutogenClient() *InMemoryAutogenClient {
	return NewInMemoryAutogenClient()
}

// GetVersion implements the Client interface
func (m *InMemoryAutogenClient) GetVersion(_ context.Context) (string, error) {
	return "1.0.0-inmemory", nil
}

// InvokeTask implements the Client interface
func (m *InMemoryAutogenClient) InvokeTask(ctx context.Context, req *autogen_client.InvokeTaskRequest) (*autogen_client.InvokeTaskResult, error) {
	// For in-memory implementation, return a basic result
	return &autogen_client.InvokeTaskResult{
		TaskResult: autogen_client.TaskResult{
			Messages: []json.RawMessage{
				json.RawMessage(fmt.Sprintf(`{"role": "assistant", "content": "Task completed: %s"}`, req.Task)),
			},
		},
	}, nil
}

// InvokeTaskStream implements the Client interface
func (m *InMemoryAutogenClient) InvokeTaskStream(ctx context.Context, req *autogen_client.InvokeTaskRequest) (<-chan *autogen_client.SseEvent, error) {
	ch := make(chan *autogen_client.SseEvent, 1)
	go func() {
		defer close(ch)
		ch <- &autogen_client.SseEvent{
			Event: "message",
			Data:  []byte(fmt.Sprintf("Task stream completed: %s", req.Task)),
		}
	}()

	return ch, nil
}

// FetchTools implements the Client interface
func (m *InMemoryAutogenClient) FetchTools(ctx context.Context, req *autogen_client.ToolServerRequest) (*autogen_client.ToolServerResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools, exists := m.toolsByServer[req.Server.Label]
	if !exists {
		return &autogen_client.ToolServerResponse{
			Tools: []*autogen_client.NamedTool{},
		}, nil
	}

	return &autogen_client.ToolServerResponse{
		Tools: tools,
	}, nil
}

// Validate implements the Client interface
func (m *InMemoryAutogenClient) Validate(ctx context.Context, req *autogen_client.ValidationRequest) (*autogen_client.ValidationResponse, error) {
	return &autogen_client.ValidationResponse{
		IsValid:  true,
		Errors:   []*autogen_client.ValidationError{},
		Warnings: []*autogen_client.ValidationError{},
	}, nil
}

// Helper method to add tools for testing purposes (not part of the interface)
func (m *InMemoryAutogenClient) AddToolsForServer(serverLabel string, tools []*autogen_client.NamedTool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.toolsByServer[serverLabel] = tools
}

func (m *InMemoryAutogenClient) ListSupportedModels(ctx context.Context) (*autogen_client.ProviderModels, error) {
	return &autogen_client.ProviderModels{
		"openai": {
			{
				Name: "gpt-4o",
			},
		},
	}, nil
}
