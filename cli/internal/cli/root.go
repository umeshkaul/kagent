package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kagent",
		Short: "kagent CLI",
		Long:  `A CLI tool to interact with kagent`,
	}

	// Add global flags
	cmd.PersistentFlags().String("api-url", "http://localhost:8081/api", "Backend API URL")
	cmd.PersistentFlags().String("ws-url", "", "WebSocket URL (optional, derived from backend URL if not set)")
	cmd.PersistentFlags().String("output", "table", "Output format (json or table)")

	// Bind flags to viper
	viper.BindPFlag("api_url", cmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag("ws_url", cmd.PersistentFlags().Lookup("ws-url"))
	viper.BindPFlag("output_format", cmd.PersistentFlags().Lookup("output"))

	// Set default values
	viper.SetDefault("output_format", "table")

	// Add commands
	cmd.AddCommand(
		newInstallCmd(),
		newConfigCmd(),
		newTeamCmd(),
		newRunCmd(),
		newVersionCmd(),
	)

	return cmd
}
