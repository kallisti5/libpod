package main

import (
	"fmt"

	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	secretRmCommand     cliconfig.SecretRmValues
	secretRmDescription = `Remove one or more existing secrets.

  By default only secrets that are not being used by any containers will be removed. To remove the secrets anyways, use the --force flag.`
	_secretRmCommand = &cobra.Command{
		Use:     "rm [flags] VOLUME [VOLUME...]",
		Aliases: []string{"remove"},
		Short:   "Remove one or more secrets",
		Long:    secretRmDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretRmCommand.InputArgs = args
			secretRmCommand.GlobalFlags = MainGlobalOpts
			secretRmCommand.Remote = remoteclient
			return secretRmCmd(&secretRmCommand)
		},
		Example: `podman secret rm myvol1 myvol2
  podman secret rm --all
  podman secret rm --force myvol`,
	}
)

func init() {
	secretRmCommand.Command = _secretRmCommand
	secretRmCommand.SetHelpTemplate(HelpTemplate())
	secretRmCommand.SetUsageTemplate(UsageTemplate())
	flags := secretRmCommand.Flags()
	flags.BoolVarP(&secretRmCommand.All, "all", "a", false, "Remove all secrets")
	flags.BoolVarP(&secretRmCommand.Force, "force", "f", false, "Remove a secret by force, even if it is being used by a container")
}

func secretRmCmd(c *cliconfig.SecretRmValues) error {
	var err error

	if (len(c.InputArgs) > 0 && c.All) || (len(c.InputArgs) < 1 && !c.All) {
		return errors.New("choose either one or more secrets or all")
	}

	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.Shutdown(false)
	deletedSecretNames, err := runtime.RemoveSecrets(getContext(), c)
	if err != nil {
		if len(deletedSecretNames) > 0 {
			printDeleteSecrets(deletedSecretNames)
			return err
		}
	}
	printDeleteSecrets(deletedSecretNames)
	return err
}

func printDeleteSecrets(secrets []string) {
	for _, v := range secrets {
		fmt.Println(v)
	}
}
