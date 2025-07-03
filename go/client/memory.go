package client

import (
	"context"
	"fmt"

	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
)

// MemoryInterface defines the memory operations
type MemoryInterface interface {
	ListMemories(ctx context.Context) ([]MemoryResponse, error)
	CreateMemory(ctx context.Context, request *CreateMemoryRequest) (*v1alpha1.Memory, error)
	GetMemory(ctx context.Context, namespace, memoryName string) (*MemoryResponse, error)
	UpdateMemory(ctx context.Context, namespace, memoryName string, request *UpdateMemoryRequest) (*v1alpha1.Memory, error)
	DeleteMemory(ctx context.Context, namespace, memoryName string) error
}

// MemoryClient handles memory-related requests
type MemoryClient struct {
	client *BaseClient
}

// NewMemoryClient creates a new memory client
func NewMemoryClient(client *BaseClient) MemoryInterface {
	return &MemoryClient{client: client}
}

// ListMemories lists all memories
func (c *MemoryClient) ListMemories(ctx context.Context) ([]MemoryResponse, error) {
	resp, err := c.client.Get(ctx, "/api/memories", "")
	if err != nil {
		return nil, err
	}

	var memories []MemoryResponse
	if err := DecodeResponse(resp, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// CreateMemory creates a new memory
func (c *MemoryClient) CreateMemory(ctx context.Context, request *CreateMemoryRequest) (*v1alpha1.Memory, error) {
	resp, err := c.client.Post(ctx, "/api/memories", request, "")
	if err != nil {
		return nil, err
	}

	var memory v1alpha1.Memory
	if err := DecodeResponse(resp, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// GetMemory retrieves a specific memory
func (c *MemoryClient) GetMemory(ctx context.Context, namespace, memoryName string) (*MemoryResponse, error) {
	path := fmt.Sprintf("/api/memories/%s/%s", namespace, memoryName)
	resp, err := c.client.Get(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var memory MemoryResponse
	if err := DecodeResponse(resp, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// UpdateMemory updates an existing memory
func (c *MemoryClient) UpdateMemory(ctx context.Context, namespace, memoryName string, request *UpdateMemoryRequest) (*v1alpha1.Memory, error) {
	path := fmt.Sprintf("/api/memories/%s/%s", namespace, memoryName)
	resp, err := c.client.Put(ctx, path, request, "")
	if err != nil {
		return nil, err
	}

	var memory v1alpha1.Memory
	if err := DecodeResponse(resp, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// DeleteMemory deletes a memory
func (c *MemoryClient) DeleteMemory(ctx context.Context, namespace, memoryName string) error {
	path := fmt.Sprintf("/api/memories/%s/%s", namespace, memoryName)
	resp, err := c.client.Delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
