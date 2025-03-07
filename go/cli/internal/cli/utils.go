package cli

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/fatih/color"
	autogen_client "github.com/kagent-dev/kagent/go/autogen/client"
	"github.com/kagent-dev/kagent/go/cli/internal/config"
)

const (
	configKey = "[config]"
	clientKey = "[client]"
)

func SetCfg(shell *ishell.Shell, cfg *config.Config) {
	shell.Set(configKey, cfg)
}

func SetClient(shell *ishell.Shell, client *autogen_client.Client) {
	shell.Set(clientKey, client)
}

func GetCfg(shell *ishell.Context) *config.Config {
	return shell.Get(configKey).(*config.Config)
}

func GetClient(shell *ishell.Context) *autogen_client.Client {
	return shell.Get(clientKey).(*autogen_client.Client)
}

func BoldBlue(s string) string {
	return color.New(color.FgBlue, color.Bold).SprintFunc()(s)
}

func BoldGreen(s string) string {
	return color.New(color.FgGreen, color.Bold).SprintFunc()(s)
}
