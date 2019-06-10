package main

import (
	"fmt"

	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/cmd/podman/shared"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	secretCreateCommand     cliconfig.SecretCreateValues
	secretCreateDescription = `If using the default driver, "local", the secret will be created on the host in the secrets directory under container storage.`

	_secretCreateCommand = &cobra.Command{
		Use:   "create [flags] [NAME]",
		Short: "Create a new secret",
		Long:  secretCreateDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretCreateCommand.InputArgs = args
			secretCreateCommand.GlobalFlags = MainGlobalOpts
			secretCreateCommand.Remote = remoteclient
			return secretCreateCmd(&secretCreateCommand)
		},
		Example: `podman secret create mysecret
  podman secret create
  podman secret create --label foo=bar myvol`,
	}
)

func init() {
	secretCreateCommand.Command = _secretCreateCommand
	secretCommand.SetHelpTemplate(HelpTemplate())
	secretCreateCommand.SetUsageTemplate(UsageTemplate())
	flags := secretCreateCommand.Flags()
	flags.StringVar(&secretCreateCommand.Driver, "driver", "", "Specify secret driver name (default local)")
	flags.StringSliceVarP(&secretCreateCommand.Label, "label", "l", []string{}, "Set metadata for a secret (default [])")
	flags.StringSliceVarP(&secretCreateCommand.Opt, "opt", "o", []string{}, "Set driver specific options (default [])")

}

func secretCreateCmd(c *cliconfig.SecretCreateValues) error {
	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.Shutdown(false)

	if len(c.InputArgs) > 1 {
		return errors.Errorf("too many arguments, create takes at most 1 argument")
	}

	labels, err := shared.GetAllLabels([]string{}, c.Label)
	if err != nil {
		return errors.Wrapf(err, "unable to process labels")
	}

	opts, err := shared.GetAllLabels([]string{}, c.Opt)
	if err != nil {
		return errors.Wrapf(err, "unable to process options")
	}

	secretName, err := runtime.CreateSecret(getContext(), c, labels, opts)
	if err == nil {
		fmt.Println(secretName)
	}
	return err
}
