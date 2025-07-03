package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// SessionInterface defines the session operations
type SessionInterface interface {
	ListSessions(ctx context.Context, userID string) ([]Session, error)
	CreateSession(ctx context.Context, request *SessionRequest) (*Session, error)
	GetSession(ctx context.Context, sessionName, userID string) (*Session, error)
	UpdateSession(ctx context.Context, request *SessionRequest) (*Session, error)
	DeleteSession(ctx context.Context, sessionName, userID string) error
	ListSessionRuns(ctx context.Context, sessionName, userID string) ([]interface{}, error)
}

// SessionClient handles session-related requests
type SessionClient struct {
	client *BaseClient
}

// NewSessionClient creates a new session client
func NewSessionClient(client *BaseClient) SessionInterface {
	return &SessionClient{client: client}
}

// ListSessions lists all sessions for a user
func (c *SessionClient) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	resp, err := c.client.Get(ctx, "/api/sessions", userID)
	if err != nil {
		return nil, err
	}

	var response StandardResponse[[]Session]
	if err := DecodeResponse(resp, &response); err != nil {
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
func (c *SessionClient) CreateSession(ctx context.Context, request *SessionRequest) (*Session, error) {
	userID := c.client.GetUserIDOrDefault(request.UserID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	request.UserID = userID

	resp, err := c.client.Post(ctx, "/api/sessions", request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Session]
	if err := DecodeResponse(resp, &response); err != nil {
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
func (c *SessionClient) GetSession(ctx context.Context, sessionName, userID string) (*Session, error) {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	path := fmt.Sprintf("/api/sessions/%s", sessionName)
	resp, err := c.client.Get(ctx, path, userID)
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Session]
	if err := DecodeResponse(resp, &response); err != nil {
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
func (c *SessionClient) UpdateSession(ctx context.Context, request *SessionRequest) (*Session, error) {
	userID := c.client.GetUserIDOrDefault(request.UserID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}
	request.UserID = userID

	resp, err := c.client.Put(ctx, "/api/sessions", request, "")
	if err != nil {
		return nil, err
	}

	var response StandardResponse[Session]
	if err := DecodeResponse(resp, &response); err != nil {
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
func (c *SessionClient) DeleteSession(ctx context.Context, sessionName, userID string) error {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return fmt.Errorf("userID is required")
	}

	path := fmt.Sprintf("/api/sessions/%s", sessionName)
	resp, err := c.client.Delete(ctx, path, userID)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ListSessionRuns lists all runs for a specific session
func (c *SessionClient) ListSessionRuns(ctx context.Context, sessionName, userID string) ([]interface{}, error) {
	userID = c.client.GetUserIDOrDefault(userID)
	if userID == "" {
		return nil, fmt.Errorf("userID is required")
	}

	path := fmt.Sprintf("/api/sessions/%s/runs", sessionName)
	resp, err := c.client.Get(ctx, path, userID)
	if err != nil {
		return nil, err
	}

	var response SessionRunsResponse
	if err := DecodeResponse(resp, &response); err != nil {
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
