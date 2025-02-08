package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kagent-dev/kagent/cli/internal/api"
	"github.com/kagent-dev/kagent/cli/internal/config"
	"github.com/kagent-dev/kagent/cli/internal/ws"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manage runs",
	}

	createCmd := &cobra.Command{
		Use:   "create [team-name] [task]",
		Short: "Create a new run",
		Args:  cobra.ExactArgs(2),
		RunE:  runRunCreate,
	}

	listAllCmd := &cobra.Command{
		Use:   "list",
		Short: "Lists all runs",
		RunE:  runListAll,
	}

	cmd.AddCommand(
		createCmd,
		listAllCmd,
		&cobra.Command{
			Use:   "get [run-id]",
			Short: "Get run details",
			Args:  cobra.ExactArgs(1),
			RunE:  runRunGet,
		},
	)

	return cmd
}

func runListAll(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	runs, err := client.ListRuns(cfg.UserID)
	if err != nil {
		return fmt.Errorf("error listing runs: %w", err)
	}

	if len(runs) == 0 {
		fmt.Println("No runs found")
		return nil
	}

	headers := []string{"ID", "CONTENT", "MESSAGES", "STATUS", "CREATED"}
	rows := make([][]string, len(runs))
	for i, run := range runs {
		// Truncate task content to first 10 characters if possible
		content := run.Task.Content
		if len(content) > 10 {
			content = content[:10] + "..."
		}

		rows[i] = []string{
			run.ID,
			content,
			strconv.Itoa(len(run.Messages)),
			run.Status,
			run.CreatedAt,
		}
	}

	return PrintOutput(runs, headers, rows)

}

func generateRandomString(prefix string, length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return prefix + string(b), nil
}

func runRunCreate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	// Get the team based on the input + userID
	team, err := client.GetTeam(args[0], cfg.UserID)
	if err != nil {
		return err
	}

	// Create a random session name
	sessionName, err := generateRandomString("session-", 5)
	if err != nil {
		return err
	}

	session, err := client.CreateSession(&api.CreateSession{
		UserID: cfg.UserID,
		// This will probably be created on the apiserver side in the future
		Name:   sessionName,
		TeamID: team.ID,
	})
	if err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		return err
	}

	run, err := client.CreateRun(&api.CreateRunRequest{
		SessionID: session.ID,
		UserID:    session.UserID,
	})
	if err != nil {
		return err
	}

	wsConfig := ws.DefaultConfig()
	wsClient, err := ws.NewClient(cfg.WSURL, run.ID, wsConfig)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket client: %v", err)
	}

	// Starting interactive mode by default
	return wsClient.StartInteractive(*team, args[1])
}

func runRunGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	run, err := client.GetRun(args[0])
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}
