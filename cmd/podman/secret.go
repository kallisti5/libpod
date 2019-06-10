package main

import (
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/spf13/cobra"
)

var secretDescription = `Secrets are created in and can be shared between containers.`

var secretCommand = cliconfig.PodmanCommand{
	Command: &cobra.Command{
		Use:   "secret",
		Short: "Manage secrets",
		Long:  secretDescription,
		RunE:  commandRunE(),
	},
}
var secretSubcommands = []*cobra.Command{
	_secretCreateCommand,
	_secretLsCommand,
	_secretRmCommand,
	_secretInspectCommand,
}

func init() {
	secretCommand.SetUsageTemplate(UsageTemplate())
	secretCommand.AddCommand(secretSubcommands...)
	rootCmd.AddCommand(secretCommand.Command)
}
