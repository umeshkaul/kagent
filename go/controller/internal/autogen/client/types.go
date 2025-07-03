package client

import (
	"encoding/json"
	"errors"
)

type TaskResult struct {
	// These are all of type Event, but we don't want to unmarshal them here
	// because we want to handle them in the caller
	Messages   []json.RawMessage `json:"messages"`
	StopReason string            `json:"stop_reason"`
}

// APIResponse is the common response wrapper for all API responses
type APIResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ProviderModels maps provider names to a list of their supported model names.
type ProviderModels map[string][]ModelInfo

// ModelInfo holds details about a specific model.
type ModelInfo struct {
	Name            string `json:"name"`
	FunctionCalling bool   `json:"function_calling"`
}

var (
	NotFoundError = errors.New("not found")
)
