package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kagent-dev/kagent/go/client"
)

func main() {
	// Create a new client with default user ID
	c := client.NewClientSet("http://localhost:8080",
		client.WithUserID("example-user"),
	)

	ctx := context.Background()

	// Test health and version
	fmt.Println("=== Health Check ===")
	if err := c.Health().Health(ctx); err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Println("✓ Server is healthy")
	}

	version, err := c.Version().GetVersion(ctx)
	if err != nil {
		log.Printf("Failed to get version: %v", err)
	} else {
		fmt.Printf("✓ Server version: %s (commit: %s)\n", version.KAgentVersion, version.GitCommit)
	}

	// List model configurations
	fmt.Println("\n=== Model Configurations ===")
	configs, err := c.ModelConfigs().ListModelConfigs(ctx)
	if err != nil {
		log.Printf("Failed to list model configs: %v", err)
	} else {
		fmt.Printf("Found %d model configurations:\n", len(configs.Data))
		for _, config := range configs.Data {
			fmt.Printf("- %s (%s, model: %s)\n", config.Ref, config.ProviderName, config.Model)
		}
	}

	// List namespaces
	fmt.Println("\n=== Namespaces ===")
	namespaces, err := c.Namespaces().ListNamespaces(ctx)
	if err != nil {
		log.Printf("Failed to list namespaces: %v", err)
	} else {
		fmt.Printf("Found %d namespaces:\n", len(namespaces))
		for _, ns := range namespaces {
			fmt.Printf("- %s (status: %s)\n", ns.Name, ns.Status)
		}
	}

	// List providers
	fmt.Println("\n=== Providers ===")
	modelProviders, err := c.Providers().ListSupportedModelProviders(ctx)
	if err != nil {
		log.Printf("Failed to list model providers: %v", err)
	} else {
		fmt.Printf("Supported model providers:\n")
		for _, provider := range modelProviders {
			fmt.Printf("- %s: required=%v, optional=%v\n",
				provider.Type, provider.RequiredParams, provider.OptionalParams)
		}
	}

	memoryProviders, err := c.Providers().ListSupportedMemoryProviders(ctx)
	if err != nil {
		log.Printf("Failed to list memory providers: %v", err)
	} else {
		fmt.Printf("Supported memory providers:\n")
		for _, provider := range memoryProviders {
			fmt.Printf("- %s: required=%v, optional=%v\n",
				provider.Type, provider.RequiredParams, provider.OptionalParams)
		}
	}

	// Demonstrate session management
	fmt.Println("\n=== Session Management ===")

	// Create a session
	sessionReq := &client.SessionRequest{
		Name:   fmt.Sprintf("example-session-%d", time.Now().Unix()),
		UserID: "example-user",
	}

	session, err := c.Sessions().CreateSession(ctx, sessionReq)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
	} else {
		fmt.Printf("✓ Created session: %s (ID: %d)\n", session.Name, session.ID)

		// List all sessions
		sessions, err := c.Sessions().ListSessions(ctx, "example-user")
		if err != nil {
			log.Printf("Failed to list sessions: %v", err)
		} else {
			fmt.Printf("✓ Total sessions for user: %d\n", len(sessions))
		}

		// Get the specific session
		retrievedSession, err := c.Sessions().GetSession(ctx, session.Name, "example-user")
		if err != nil {
			log.Printf("Failed to get session: %v", err)
		} else {
			fmt.Printf("✓ Retrieved session: %s\n", retrievedSession.Name)
		}

		// List session runs (should be empty for new session)
		runs, err := c.Sessions().ListSessionRuns(ctx, session.Name, "example-user")
		if err != nil {
			log.Printf("Failed to list session runs: %v", err)
		} else {
			fmt.Printf("✓ Session runs: %d\n", len(runs))
		}

		// Clean up - delete the session
		err = c.Sessions().DeleteSession(ctx, session.Name, "example-user")
		if err != nil {
			log.Printf("Failed to delete session: %v", err)
		} else {
			fmt.Printf("✓ Deleted session: %s\n", session.Name)
		}
	}

	// List tools
	fmt.Println("\n=== Tools ===")
	tools, err := c.Tools().ListTools(ctx, "example-user")
	if err != nil {
		log.Printf("Failed to list tools: %v", err)
	} else {
		fmt.Printf("Found %d tools for user:\n", len(tools))
		for _, tool := range tools {
			fmt.Printf("- %s (ID: %d)\n", tool.Name, tool.ID)
		}
	}

	// List tool servers
	fmt.Println("\n=== Tool Servers ===")
	toolServers, err := c.ToolServers().ListToolServers(ctx)
	if err != nil {
		log.Printf("Failed to list tool servers: %v", err)
	} else {
		fmt.Printf("Found %d tool servers:\n", len(toolServers))
		for _, ts := range toolServers {
			fmt.Printf("- %s (discovered tools: %d)\n", ts.Ref, len(ts.DiscoveredTools))
		}
	}

	// List teams
	fmt.Println("\n=== Teams ===")
	teams, err := c.Teams().ListTeams(ctx, "example-user")
	if err != nil {
		log.Printf("Failed to list teams: %v", err)
	} else {
		fmt.Printf("Found %d teams for user:\n", len(teams))
		for _, team := range teams {
			fmt.Printf("- %s (ID: %d)\n", team.Name, team.ID)
		}
	}

	// List memories
	fmt.Println("\n=== Memories ===")
	memories, err := c.Memories().ListMemories(ctx)
	if err != nil {
		log.Printf("Failed to list memories: %v", err)
	} else {
		fmt.Printf("Found %d memories:\n", len(memories))
		for _, memory := range memories {
			fmt.Printf("- %s (%s)\n", memory.Ref, memory.ProviderName)
		}
	}

	// List feedback
	fmt.Println("\n=== Feedback ===")
	feedback, err := c.Feedback().ListFeedback(ctx, "example-user")
	if err != nil {
		log.Printf("Failed to list feedback: %v", err)
	} else {
		fmt.Printf("Found %d feedback entries for user:\n", len(feedback))
		for _, fb := range feedback {
			positivity := "negative"
			if fb.IsPositive {
				positivity = "positive"
			}
			fmt.Printf("- %s feedback: %s\n", positivity, fb.FeedbackText)
		}
	}

	// Demonstrate error handling
	fmt.Println("\n=== Error Handling ===")
	_, err = c.ModelConfigs().GetModelConfig(ctx, "nonexistent", "config")
	if err != nil {
		if clientErr, ok := err.(*client.ClientError); ok {
			fmt.Printf("✓ Expected error: HTTP %d - %s\n", clientErr.StatusCode, clientErr.Message)
		} else {
			fmt.Printf("✓ Expected error: %v\n", err)
		}
	}

	fmt.Println("\n=== Example Complete ===")
}
