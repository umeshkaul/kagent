package client

import "github.com/kagent-dev/kagent/go/autogen/api"

type InvokeTaskRequest struct {
	Task       string         `json:"task"`
	TeamConfig *api.Component `json:"team_config"`
}

type InvokeTaskResult struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Data    struct {
		Usage      string  `json:"usage"`
		Duration   float64 `json:"duration"`
		TaskResult struct {
			StopReason string                   `json:"stop_reason"`
			Messages   []map[string]interface{} `json:"messages"`
		} `json:"task_result"`
	} `json:"data"`
}

func (c *Client) InvokeTask(req *InvokeTaskRequest) (*InvokeTaskResult, error) {
	var invoke InvokeTaskResult
	err := c.doRequest("POST", "/invoke", req, &invoke)
	return &invoke, err
}
