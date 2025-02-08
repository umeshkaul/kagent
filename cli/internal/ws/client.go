package ws

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kagent-dev/kagent/cli/internal/api"
)

// Config holds the WebSocket client configuration
type Config struct {
	Origin string // WebSocket origin header
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Origin: "http://localhost:8000",
	}
}

// Client handles the WebSocket connection and message processing
type Client struct {
	conn   *websocket.Conn
	done   chan struct{}
	config Config
}

// NewClient creates a new WebSocket client and establishes connection
func NewClient(wsURL string, runID string, config Config) (*Client, error) {
	// Set the required headers for the WebSocket connection
	headers := http.Header{}
	headers.Add("Origin", config.Origin)

	// Create dialer with debug logging
	dialer := websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}

	conn, _, err := dialer.Dial(wsURL+"/runs/"+runID, headers)
	if err != nil {
		return nil, fmt.Errorf("websocket connection failed: %v", err)
	}

	return &Client{
		conn:   conn,
		done:   make(chan struct{}),
		config: config,
	}, nil
}

// StartInteractive initiates the interactive session with the server
func (c *Client) StartInteractive(team api.Team, task string) error {
	defer c.conn.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	inputTimeout := make(chan struct{})

	// Send initial start message
	startMsg := StartMessage{
		Type:       MessageTypeStart,
		Task:       task,
		TeamConfig: team.Component,
	}

	if err := c.conn.WriteJSON(startMsg); err != nil {
		return fmt.Errorf("failed to send start message: %v", err)
	}

	go c.handleMessages(inputTimeout)

	select {
	case <-interrupt:
		fmt.Println("\nReceived interrupt signal. Closing connection...")
		stopMsg := StopMessage{
			Type:   MessageTypeStop,
			Reason: "Cancelled by user",
		}
		if err := c.conn.WriteJSON(stopMsg); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending stop message: %v\n", err)
		}
		select {
		case <-c.done:
		case <-time.After(time.Second):
		}
		return nil

	case <-inputTimeout:
		fmt.Println("\nInput timeout exceeded. Stopping task...")
		return fmt.Errorf("input timeout exceeded")

	case <-c.done:
		return nil
	}
}

func getMessageContentType(data json.RawMessage) (ContentType, error) {
	mapStructure := &map[string]string{}
	if err := json.Unmarshal(data, mapStructure); err != nil {
		return "", fmt.Errorf("error parsing message data: %v", err)
	}

	typeStr := (*mapStructure)["type"]
	return ContentType(typeStr), nil
}

func (c *Client) handleMessages(inputTimeout chan struct{}) {
	defer close(c.done)

	for {
		var msg BaseWebSocketMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading message: %v\n", err)
			return
		}
		switch msg.Type {
		case MessageTypeError:
			fmt.Fprintf(os.Stderr, "Error: %s\n", msg.Error)
			return

		case MessageTypeMessage:
			mapStructure := &map[string]string{}
			_ = json.Unmarshal(msg.Data, mapStructure)
			typeStr := (*mapStructure)["type"]
			contentType := ContentType(typeStr)

			switch contentType {
			case ContentTypeText:
				var textMessage TextMessage
				if err := json.Unmarshal(msg.Data, &textMessage); err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing message data: %v\n", err)
					continue
				}
				fmt.Printf("%s: %s\n", textMessage.Source, textMessage.Content)
			case ContentTypeToolCallRequest:
				var toolCallRequest ToolCallRequest
				if err := json.Unmarshal(msg.Data, &toolCallRequest); err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing message data: %v\n", err)
					continue
				}
				fmt.Printf("%s (function call request) \n", toolCallRequest.Source)
				for _, functionCall := range toolCallRequest.Content {
					fmt.Printf("  %s(%s)\n", functionCall.Name, functionCall.Arguments)
				}

			case ContentTypeToolCallExecution:
				var toolCallExecution ToolCallExecution
				if err := json.Unmarshal(msg.Data, &toolCallExecution); err != nil {
					fmt.Fprintf(os.Stderr, "Error parsing message data: %v\n", err)
					continue
				}
				fmt.Printf("%s (function call execution)\n", toolCallExecution.Source)
				for _, functionExecution := range toolCallExecution.Content {
					fmt.Printf("%s\n", functionExecution.Content)
				}
			}

		case MessageTypeInputRequest:
			go c.handleInputTimeout(inputTimeout)
			if err := c.handleUserInput(msg.Data); err != nil {
				fmt.Fprintf(os.Stderr, "Error handling input: %v\n", err)
				return
			}

		case MessageTypeResult, MessageTypeCompletion:
			var msgResult CompletionMessage
			if err := json.Unmarshal(msg.Data, &msgResult); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing message data: %v\n", err)
				continue
			}

			fmt.Printf("\n(%s) Task completed:\n%s", msgResult.Status, msgResult.Data)
		}

	}
}

func (c *Client) handleInputTimeout(inputTimeout chan struct{}) {
	timer := time.NewTimer(InputTimeoutDuration)
	select {
	case <-timer.C:
		close(inputTimeout)
		stopMsg := StopMessage{
			Type:   MessageTypeStop,
			Reason: "Input timeout",
			Code:   4000,
		}
		c.conn.WriteJSON(stopMsg)
	case <-c.done:
		timer.Stop()
	}
}

func (c *Client) handleUserInput(msg json.RawMessage) error {
	var inputRequest InputRequestMessage
	if err := json.Unmarshal(msg, &inputRequest); err != nil {
		return fmt.Errorf("error parsing input request: %v", err)
	}
	fmt.Printf("%s: %s\n", inputRequest.Source, inputRequest.Content)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := scanner.Text()
		inputMsg := InputResponseMessage{
			Type:     MessageTypeInputResponse,
			Response: response,
		}
		return c.conn.WriteJSON(inputMsg)
	}
	return fmt.Errorf("failed to read user input")
}
