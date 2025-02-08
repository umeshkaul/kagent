package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	WSURL      string
	HTTPClient *http.Client
}

func (c *Client) GetVersion() (string, error) {
	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Version string `json:"version"`
		} `json:"data"`
	}

	err := c.doRequest("GET", "/version", nil, &result)
	if err != nil {
		return "", err
	}

	if !result.Status {
		return "", fmt.Errorf("api error: %s", result.Message)
	}

	return result.Data.Version, nil
}

func NewClient(baseURL, wsURL string) *Client {
	// Ensure baseURL doesn't end with a slash
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		BaseURL: baseURL,
		WSURL:   wsURL,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Ensure path starts with a slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := c.BaseURL + path

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequest(method, url, bodyReader)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	// Decode into APIResponse first
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	// Check response status
	if !apiResp.Status {
		return fmt.Errorf("api error: %s", apiResp.Message)
	}

	// If caller wants the result, marshal the Data field into their result type
	if result != nil {
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("error unmarshaling into result: %w", err)
		}
	}

	return nil
}
