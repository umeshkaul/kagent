package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		options     []ClientOption
		expectedURL string
	}{
		{
			name:        "basic client",
			baseURL:     "http://localhost:8080",
			options:     nil,
			expectedURL: "http://localhost:8080",
		},
		{
			name:        "client with trailing slash",
			baseURL:     "http://localhost:8080/",
			options:     nil,
			expectedURL: "http://localhost:8080",
		},
		{
			name:    "client with user ID",
			baseURL: "http://localhost:8080",
			options: []ClientOption{WithUserID("test-user")},
		},
		{
			name:    "client with custom HTTP client",
			baseURL: "http://localhost:8080",
			options: []ClientOption{WithHTTPClient(&http.Client{Timeout: 60 * time.Second})},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.baseURL, tt.options...)

			if client.BaseURL != tt.expectedURL {
				t.Errorf("expected BaseURL %s, got %s", tt.expectedURL, client.BaseURL)
			}

			if client.HTTPClient == nil {
				t.Error("HTTPClient should not be nil")
			}
		})
	}
}

func TestClientHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()

	err := client.Health(ctx)
	if err != nil {
		t.Errorf("Health() returned unexpected error: %v", err)
	}
}

func TestClientGetVersion(t *testing.T) {
	expectedVersion := VersionResponse{
		KAgentVersion: "1.0.0",
		GitCommit:     "abc123",
		BuildDate:     "2024-01-01",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version" {
			t.Errorf("expected path /version, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedVersion)
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()

	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatalf("GetVersion() returned unexpected error: %v", err)
	}

	if version.KAgentVersion != expectedVersion.KAgentVersion {
		t.Errorf("expected version %s, got %s", expectedVersion.KAgentVersion, version.KAgentVersion)
	}
}

func TestClientListModelConfigs(t *testing.T) {
	expectedConfigs := []ModelConfigResponse{
		{
			Ref:          "default/test-config",
			ProviderName: "OpenAI",
			Model:        "gpt-4",
			ModelParams:  map[string]interface{}{"temperature": "0.7"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/modelconfigs" {
			t.Errorf("expected path /api/modelconfigs, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedConfigs)
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()

	configs, err := client.ListModelConfigs(ctx)
	if err != nil {
		t.Fatalf("ListModelConfigs() returned unexpected error: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(configs))
	}

	if configs[0].Ref != expectedConfigs[0].Ref {
		t.Errorf("expected ref %s, got %s", expectedConfigs[0].Ref, configs[0].Ref)
	}
}

func TestClientCreateSession(t *testing.T) {
	expectedSession := Session{
		ID:     1,
		Name:   "test-session",
		UserID: "test-user",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sessions" {
			t.Errorf("expected path /api/sessions, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var sessionReq SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&sessionReq); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if sessionReq.Name != "test-session" {
			t.Errorf("expected session name test-session, got %s", sessionReq.Name)
		}

		response := NewResponse(expectedSession, "Session created successfully", false)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()

	sessionReq := &SessionRequest{
		Name:   "test-session",
		UserID: "test-user",
	}

	session, err := client.CreateSession(ctx, sessionReq)
	if err != nil {
		t.Fatalf("CreateSession() returned unexpected error: %v", err)
	}

	if session.Name != expectedSession.Name {
		t.Errorf("expected session name %s, got %s", expectedSession.Name, session.Name)
	}
}

func TestClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIError{Error: "Resource not found"})
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()

	_, err := client.GetModelConfig(ctx, "nonexistent", "config")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	clientErr, ok := err.(*ClientError)
	if !ok {
		t.Fatalf("expected ClientError, got %T", err)
	}

	if clientErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, clientErr.StatusCode)
	}

	if clientErr.Message != "Resource not found" {
		t.Errorf("expected message 'Resource not found', got %s", clientErr.Message)
	}
}

func TestClientUserIDParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID != "test-user" {
			t.Errorf("expected user_id test-user, got %s", userID)
		}

		response := NewResponse([]Session{}, "Sessions fetched successfully", false)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(server.URL, WithUserID("test-user"))
	ctx := context.Background()

	_, err := client.ListSessions(ctx, "")
	if err != nil {
		t.Fatalf("ListSessions() returned unexpected error: %v", err)
	}
}

func TestClientAddUserIDParam(t *testing.T) {
	client := New("http://localhost:8080")

	tests := []struct {
		name        string
		urlStr      string
		userID      string
		expected    string
		expectError bool
	}{
		{
			name:     "add user ID to URL without query params",
			urlStr:   "http://localhost:8080/api/sessions",
			userID:   "test-user",
			expected: "http://localhost:8080/api/sessions?user_id=test-user",
		},
		{
			name:     "add user ID to URL with existing query params",
			urlStr:   "http://localhost:8080/api/sessions?param=value",
			userID:   "test-user",
			expected: "http://localhost:8080/api/sessions?param=value&user_id=test-user",
		},
		{
			name:     "empty user ID",
			urlStr:   "http://localhost:8080/api/sessions",
			userID:   "",
			expected: "http://localhost:8080/api/sessions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := client.addUserIDParam(tt.urlStr, tt.userID)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
