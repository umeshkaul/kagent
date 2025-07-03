package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
)

// Client represents the KAgent HTTP client
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

// New creates a new KAgent HTTP client
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

// Error handling

// ClientError represents a client-side error
type ClientError struct {
	StatusCode int
	Message    string
	Body       string
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// HTTP helper methods

func (c *Client) buildURL(path string) string {
	return c.BaseURL + path
}

func (c *Client) addUserIDParam(urlStr string, userID string) (string, error) {
	if userID == "" {
		return urlStr, nil
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("user_id", userID)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, userID string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	urlStr := c.buildURL(path)
	if userID != "" {
		var err error
		urlStr, err = c.addUserIDParam(urlStr, userID)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var apiErr APIError
		if json.Unmarshal(bodyBytes, &apiErr) == nil && apiErr.Error != "" {
			return nil, &ClientError{
				StatusCode: resp.StatusCode,
				Message:    apiErr.Error,
				Body:       string(bodyBytes),
			}
		}

		return nil, &ClientError{
			StatusCode: resp.StatusCode,
			Message:    "Request failed",
			Body:       string(bodyBytes),
		}
	}

	return resp, nil
}

func (c *Client) get(ctx context.Context, path string, userID string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil, userID)
}

func (c *Client) post(ctx context.Context, path string, body interface{}, userID string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body, userID)
}

func (c *Client) put(ctx context.Context, path string, body interface{}, userID string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body, userID)
}

func (c *Client) delete(ctx context.Context, path string, userID string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil, userID)
}

func decodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// Health and Version methods

// Health checks if the server is healthy
func (c *Client) Health(ctx context.Context) error {
	resp, err := c.get(ctx, "/health", "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// GetVersion retrieves version information
func (c *Client) GetVersion(ctx context.Context) (*VersionResponse, error) {
	resp, err := c.get(ctx, "/version", "")
	if err != nil {
		return nil, err
	}

	var version VersionResponse
	if err := decodeResponse(resp, &version); err != nil {
		return nil, err
	}

	return &version, nil
}

// Model Configuration methods

// ListModelConfigs lists all model configurations
func (c *Client) ListModelConfigs(ctx context.Context) ([]ModelConfigResponse, error) {
	resp, err := c.get(ctx, "/api/modelconfigs", "")
	if err != nil {
		return nil, err
	}

	var configs []ModelConfigResponse
	if err := decodeResponse(resp, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

// GetModelConfig retrieves a specific model configuration
func (c *Client) GetModelConfig(ctx context.Context, namespace, configName string) (*ModelConfigResponse, error) {
	path := fmt.Sprintf("/api/modelconfigs/%s/%s", namespace, configName)
	resp, err := c.get(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var config ModelConfigResponse
	if err := decodeResponse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// CreateModelConfig creates a new model configuration
func (c *Client) CreateModelConfig(ctx context.Context, request *CreateModelConfigRequest) (*v1alpha1.ModelConfig, error) {
	resp, err := c.post(ctx, "/api/modelconfigs", request, "")
	if err != nil {
		return nil, err
	}

	var config v1alpha1.ModelConfig
	if err := decodeResponse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateModelConfig updates an existing model configuration
func (c *Client) UpdateModelConfig(ctx context.Context, namespace, configName string, request *UpdateModelConfigRequest) (*ModelConfigResponse, error) {
	path := fmt.Sprintf("/api/modelconfigs/%s/%s", namespace, configName)
	resp, err := c.put(ctx, path, request, "")
	if err != nil {
		return nil, err
	}

	var config ModelConfigResponse
	if err := decodeResponse(resp, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DeleteModelConfig deletes a model configuration
func (c *Client) DeleteModelConfig(ctx context.Context, namespace, configName string) error {
	path := fmt.Sprintf("/api/modelconfigs/%s/%s", namespace, configName)
	resp, err := c.delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Session methods

// ListSessions lists all sessions for a user
func (c *Client) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.get(ctx, "/api/sessions", userID)
	if err != nil {
		return nil, err
	}

	var response StandardResponse[[]Session]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	sessionsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	if err := json.Unmarshal(sessionsData, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// CreateSession creates a new session
func (c *Client) CreateSession(ctx context.Context, request *SessionRequest) (*Session, error) {
	userID := request.UserID
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	request.UserID = userID

	resp, err := c.post(ctx, "/api/sessions", request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Session]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	sessionData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// GetSession retrieves a specific session
func (c *Client) GetSession(ctx context.Context, sessionName, userID string) (*Session, error) {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	path := fmt.Sprintf("/api/sessions/%s", sessionName)
	resp, err := c.get(ctx, path, userID)
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Session]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	sessionData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (c *Client) UpdateSession(ctx context.Context, request *SessionRequest) (*Session, error) {
	userID := request.UserID
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	request.UserID = userID

	resp, err := c.put(ctx, "/api/sessions", request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Session]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	sessionData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// DeleteSession deletes a session
func (c *Client) DeleteSession(ctx context.Context, sessionName, userID string) error {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	path := fmt.Sprintf("/api/sessions/%s", sessionName)
	resp, err := c.delete(ctx, path, userID)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ListSessionRuns lists all runs for a specific session
func (c *Client) ListSessionRuns(ctx context.Context, sessionName, userID string) ([]interface{}, error) {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	path := fmt.Sprintf("/api/sessions/%s/runs", sessionName)
	resp, err := c.get(ctx, path, userID)
	if err != nil {
		return nil, err
	}

	var response SessionRunsResponse
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	runData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var runsData SessionRunsData
	if err := json.Unmarshal(runData, &runsData); err != nil {
		return nil, err
	}

	return runsData.Runs, nil
}

// Tool methods

// ListTools lists all tools for a user
func (c *Client) ListTools(ctx context.Context, userID string) ([]Tool, error) {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.get(ctx, "/api/tools", userID)
	if err != nil {
		return nil, err
	}

	var tools []Tool
	if err := decodeResponse(resp, &tools); err != nil {
		return nil, err
	}

	return tools, nil
}

// ToolServer methods

// ListToolServers lists all tool servers
func (c *Client) ListToolServers(ctx context.Context) ([]ToolServerResponse, error) {
	resp, err := c.get(ctx, "/api/toolservers", "")
	if err != nil {
		return nil, err
	}

	var toolServers []ToolServerResponse
	if err := decodeResponse(resp, &toolServers); err != nil {
		return nil, err
	}

	return toolServers, nil
}

// CreateToolServer creates a new tool server
func (c *Client) CreateToolServer(ctx context.Context, toolServer *v1alpha1.ToolServer) (*v1alpha1.ToolServer, error) {
	resp, err := c.post(ctx, "/api/toolservers", toolServer, "")
	if err != nil {
		return nil, err
	}

	var createdToolServer v1alpha1.ToolServer
	if err := decodeResponse(resp, &createdToolServer); err != nil {
		return nil, err
	}

	return &createdToolServer, nil
}

// DeleteToolServer deletes a tool server
func (c *Client) DeleteToolServer(ctx context.Context, namespace, toolServerName string) error {
	path := fmt.Sprintf("/api/toolservers/%s/%s", namespace, toolServerName)
	resp, err := c.delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Team methods

// ListTeams lists all teams for a user
func (c *Client) ListTeams(ctx context.Context, userID string) ([]Team, error) {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.get(ctx, "/api/teams", userID)
	if err != nil {
		return nil, err
	}

	var response StandardResponse[[]Team]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamsData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var teams []Team
	if err := json.Unmarshal(teamsData, &teams); err != nil {
		return nil, err
	}

	return teams, nil
}

// CreateTeam creates a new team
func (c *Client) CreateTeam(ctx context.Context, request *TeamRequest) (*Team, error) {
	resp, err := c.post(ctx, "/api/teams", request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Team]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(teamData, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// GetTeam retrieves a specific team
func (c *Client) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	path := fmt.Sprintf("/api/teams/%s", teamID)
	resp, err := c.get(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Team]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(teamData, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// UpdateTeam updates an existing team
func (c *Client) UpdateTeam(ctx context.Context, teamID string, request *TeamRequest) (*Team, error) {
	path := fmt.Sprintf("/api/teams/%s", teamID)
	resp, err := c.put(ctx, path, request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Team]
	if err := decodeResponse(resp, &response); err != nil {
		return nil, err
	}

	teamData, err := json.Marshal(response.Data)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(teamData, &team); err != nil {
		return nil, err
	}

	return &team, nil
}

// DeleteTeam deletes a team
func (c *Client) DeleteTeam(ctx context.Context, teamID string) error {
	path := fmt.Sprintf("/api/teams/%s", teamID)
	resp, err := c.delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Provider methods

// ListSupportedModelProviders lists all supported model providers
func (c *Client) ListSupportedModelProviders(ctx context.Context) ([]ProviderInfo, error) {
	resp, err := c.get(ctx, "/api/providers/models", "")
	if err != nil {
		return nil, err
	}

	var providers []ProviderInfo
	if err := decodeResponse(resp, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

// ListSupportedMemoryProviders lists all supported memory providers
func (c *Client) ListSupportedMemoryProviders(ctx context.Context) ([]ProviderInfo, error) {
	resp, err := c.get(ctx, "/api/providers/memories", "")
	if err != nil {
		return nil, err
	}

	var providers []ProviderInfo
	if err := decodeResponse(resp, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}

// Model methods

// ListSupportedModels lists all supported models
func (c *Client) ListSupportedModels(ctx context.Context) (interface{}, error) {
	resp, err := c.get(ctx, "/api/models", "")
	if err != nil {
		return nil, err
	}

	var models interface{}
	if err := decodeResponse(resp, &models); err != nil {
		return nil, err
	}

	return models, nil
}

// Memory methods

// ListMemories lists all memories
func (c *Client) ListMemories(ctx context.Context) ([]MemoryResponse, error) {
	resp, err := c.get(ctx, "/api/memories", "")
	if err != nil {
		return nil, err
	}

	var memories []MemoryResponse
	if err := decodeResponse(resp, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// CreateMemory creates a new memory
func (c *Client) CreateMemory(ctx context.Context, request *CreateMemoryRequest) (*v1alpha1.Memory, error) {
	resp, err := c.post(ctx, "/api/memories", request, "")
	if err != nil {
		return nil, err
	}

	var memory v1alpha1.Memory
	if err := decodeResponse(resp, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// GetMemory retrieves a specific memory
func (c *Client) GetMemory(ctx context.Context, namespace, memoryName string) (*MemoryResponse, error) {
	path := fmt.Sprintf("/api/memories/%s/%s", namespace, memoryName)
	resp, err := c.get(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var memory MemoryResponse
	if err := decodeResponse(resp, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// UpdateMemory updates an existing memory
func (c *Client) UpdateMemory(ctx context.Context, namespace, memoryName string, request *UpdateMemoryRequest) (*v1alpha1.Memory, error) {
	path := fmt.Sprintf("/api/memories/%s/%s", namespace, memoryName)
	resp, err := c.put(ctx, path, request, "")
	if err != nil {
		return nil, err
	}

	var memory v1alpha1.Memory
	if err := decodeResponse(resp, &memory); err != nil {
		return nil, err
	}

	return &memory, nil
}

// DeleteMemory deletes a memory
func (c *Client) DeleteMemory(ctx context.Context, namespace, memoryName string) error {
	path := fmt.Sprintf("/api/memories/%s/%s", namespace, memoryName)
	resp, err := c.delete(ctx, path, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// Namespace methods

// ListNamespaces lists all namespaces
func (c *Client) ListNamespaces(ctx context.Context) ([]NamespaceResponse, error) {
	resp, err := c.get(ctx, "/api/namespaces", "")
	if err != nil {
		return nil, err
	}

	var namespaces []NamespaceResponse
	if err := decodeResponse(resp, &namespaces); err != nil {
		return nil, err
	}

	return namespaces, nil
}

// Feedback methods

// CreateFeedback creates new feedback
func (c *Client) CreateFeedback(ctx context.Context, feedback *Feedback, userID string) error {
	if userID == "" {
		userID = c.UserID
	}
	feedback.UserID = userID

	resp, err := c.post(ctx, "/api/feedback", feedback, "")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ListFeedback lists all feedback for a user
func (c *Client) ListFeedback(ctx context.Context, userID string) ([]Feedback, error) {
	if userID == "" {
		userID = c.UserID
	}
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.get(ctx, "/api/feedback", userID)
	if err != nil {
		return nil, err
	}

	var feedback []Feedback
	if err := decodeResponse(resp, &feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}
