package main

import (
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	secretInspectCommand     cliconfig.SecretInspectValues
	secretInspectDescription = `Display detailed information on one or more secrets.

  Use a Go template to change the format from JSON.`
	_secretInspectCommand = &cobra.Command{
		Use:   "inspect [flags] VOLUME [VOLUME...]",
		Short: "Display detailed information on one or more secrets",
		Long:  secretInspectDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretInspectCommand.InputArgs = args
			secretInspectCommand.GlobalFlags = MainGlobalOpts
			secretInspectCommand.Remote = remoteclient
			return secretInspectCmd(&secretInspectCommand)
		},
		Example: `podman secret inspect myvol
  podman secret inspect --all
  podman secret inspect --format "{{.Driver}} {{.Scope}}" myvol`,
	}
)

func init() {
	secretInspectCommand.Command = _secretInspectCommand
	secretInspectCommand.SetHelpTemplate(HelpTemplate())
	secretInspectCommand.SetUsageTemplate(UsageTemplate())
	flags := secretInspectCommand.Flags()
	flags.BoolVarP(&secretInspectCommand.All, "all", "a", false, "Inspect all secrets")
	flags.StringVarP(&secretInspectCommand.Format, "format", "f", "json", "Format secret output using Go template")

}

func secretInspectCmd(c *cliconfig.SecretInspectValues) error {
	if (c.All && len(c.InputArgs) > 0) || (!c.All && len(c.InputArgs) < 1) {
		return errors.New("provide one or more secret names or use --all")
	}

	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.Shutdown(false)

	vols, err := runtime.InspectSecrets(getContext(), c)
	if err != nil {
		return err
	}
	return generateVolLsOutput(vols, secretLsOptions{Format: c.Format})
}
