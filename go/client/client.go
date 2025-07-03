package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
)

// Error handling types

// ClientError represents a client-side error
type ClientError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Legacy client types for backward compatibility

// Client represents the KAgent HTTP client (legacy)
// Deprecated: Use ClientSetInterface instead
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserID     string // Default user ID for requests that require it
}

// ClientOption represents a configuration option for the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithUserID sets a default user ID for requests
func WithUserID(userID string) ClientOption {
	return func(c *Client) {
		c.UserID = userID
	}
}

// New creates a new KAgent HTTP client (legacy)
// Deprecated: Use NewClientSet instead for the modern interface-based client
func New(baseURL string, options ...ClientOption) *Client {
	client := &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// Legacy method wrappers for backward compatibility
// These methods delegate to the new clientset interface

func (c *Client) clientSet() ClientSetInterface {
	return NewClientSet(c.BaseURL, WithHTTPClient(c.HTTPClient), WithUserID(c.UserID))
}

// Health methods
func (c *Client) Health(ctx context.Context) error {
	return c.clientSet().Health().Health(ctx)
}

// Version methods
func (c *Client) GetVersion(ctx context.Context) (*VersionResponse, error) {
	return c.clientSet().Version().GetVersion(ctx)
}

// Model config methods
func (c *Client) ListModelConfigs(ctx context.Context) (*StandardResponse[[]ModelConfigResponse], error) {
	return c.clientSet().ModelConfigs().ListModelConfigs(ctx)
}

func (c *Client) GetModelConfig(ctx context.Context, namespace, name string) (*ModelConfigResponse, error) {
	return c.clientSet().ModelConfigs().GetModelConfig(ctx, namespace, name)
}

func (c *Client) CreateModelConfig(ctx context.Context, request *CreateModelConfigRequest) (*v1alpha1.ModelConfig, error) {
	return c.clientSet().ModelConfigs().CreateModelConfig(ctx, request)
}

func (c *Client) UpdateModelConfig(ctx context.Context, namespace, name string, request *UpdateModelConfigRequest) (*ModelConfigResponse, error) {
	return c.clientSet().ModelConfigs().UpdateModelConfig(ctx, namespace, name, request)
}

func (c *Client) DeleteModelConfig(ctx context.Context, namespace, name string) error {
	return c.clientSet().ModelConfigs().DeleteModelConfig(ctx, namespace, name)
}

// Session methods
func (c *Client) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	return c.clientSet().Sessions().ListSessions(ctx, userID)
}

func (c *Client) CreateSession(ctx context.Context, request *SessionRequest) (*Session, error) {
	return c.clientSet().Sessions().CreateSession(ctx, request)
}

func (c *Client) GetSession(ctx context.Context, sessionName, userID string) (*Session, error) {
	return c.clientSet().Sessions().GetSession(ctx, sessionName, userID)
}

func (c *Client) UpdateSession(ctx context.Context, request *SessionRequest) (*Session, error) {
	return c.clientSet().Sessions().UpdateSession(ctx, request)
}

func (c *Client) DeleteSession(ctx context.Context, sessionName, userID string) error {
	return c.clientSet().Sessions().DeleteSession(ctx, sessionName, userID)
}

func (c *Client) ListSessionRuns(ctx context.Context, sessionName, userID string) ([]interface{}, error) {
	return c.clientSet().Sessions().ListSessionRuns(ctx, sessionName, userID)
}

// Helper method for backward compatibility with tests
func (c *Client) addUserIDParam(urlStr string, userID string) (string, error) {
	baseClient := NewBaseClient(c.BaseURL, c.HTTPClient, c.UserID)
	return baseClient.addUserIDParam(urlStr, userID)
}
