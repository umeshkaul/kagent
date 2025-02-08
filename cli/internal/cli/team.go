package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kagent-dev/kagent/cli/internal/api"
	"github.com/kagent-dev/kagent/cli/internal/config"
	"github.com/spf13/cobra"
)

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage teams",
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new team",
		RunE:  runTeamCreate,
	}

	createCmd.Flags().StringP("file", "f", "", "Team configuration file")
	createCmd.MarkFlagRequired("file")

	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all teams",
			RunE:  runTeamList,
		},
		createCmd,
		&cobra.Command{
			Use:   "get [name]",
			Short: "Get team details",
			Args:  cobra.ExactArgs(1),
			RunE:  runTeamGet,
		},
		&cobra.Command{
			Use:   "delete [team-id]",
			Short: "Delete a team",
			Args:  cobra.ExactArgs(1),
			RunE:  runTeamDelete,
		},
	)

	return cmd
}

func runTeamList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	teams, err := client.ListTeams(cfg.UserID)
	if err != nil {
		return err
	}
	if len(teams) == 0 {
		fmt.Println("No teams found")
		return nil
	}

	// Prepare table data
	headers := []string{"ID", "NAME", "CREATED"}
	rows := make([][]string, len(teams))
	for i, team := range teams {
		rows[i] = []string{
			fmt.Sprintf("%d", team.ID),
			*team.Component.Label,
			team.CreatedAt,
		}
	}

	return PrintOutput(teams, headers, rows)
}

func runTeamCreate(cmd *cobra.Command, args []string) error {
	filePath, err := cmd.Flags().GetString("file")
	if err != nil {
		return fmt.Errorf("error getting file flag: %w", err)
	}

	jsonFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	var team api.CreateTeamRequest
	if err := json.Unmarshal(jsonFile, &team); err != nil {
		return fmt.Errorf("error parsing YAML: %w", err)
	}

	cfg, err := config.Get()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	if err = client.CreateTeam(&team); err != nil {
		return fmt.Errorf("error creating team: %w", err)
	}

	fmt.Println("Team created successfully")
	return nil
}

func runTeamGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	team, err := client.GetTeam(args[0], cfg.UserID)
	if err != nil {
		return err
	}

	if team == nil {
		fmt.Println("Team not found")
		return nil
	}

	output, err := json.MarshalIndent(team, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

func runTeamDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Get()
	if err != nil {
		return err
	}

	teamID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid team ID: %s", args[0])
	}

	client := api.NewClient(cfg.APIURL, cfg.WSURL)
	if err := client.DeleteTeam(teamID, cfg.UserID); err != nil {
		return err
	}

	fmt.Printf("Team %d deleted\n", teamID)
	return nil
}
