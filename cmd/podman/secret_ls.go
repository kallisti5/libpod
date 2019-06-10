package main

import (
	"reflect"
	"strings"

	"github.com/containers/buildah/pkg/formats"
	"github.com/containers/libpod/cmd/podman/cliconfig"
	"github.com/containers/libpod/pkg/adapter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// secretOptions is the "ls" command options
type secretLsOptions struct {
	Format string
	Quiet  bool
}

// secretLsTemplateParams is the template parameters to list the secrets
type secretLsTemplateParams struct {
	Name       string
	Labels     string
	MountPoint string
	Driver     string
	Options    string
	Scope      string
}

// secretLsJSONParams is the JSON parameters to list the secrets
type secretLsJSONParams struct {
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	MountPoint string            `json:"mountPoint"`
	Driver     string            `json:"driver"`
	Options    map[string]string `json:"options"`
	Scope      string            `json:"scope"`
}

var (
	secretLsCommand cliconfig.SecretLsValues

	secretLsDescription = `
podman secret ls

List all available secrets. The output of the secrets can be filtered
and the output format can be changed to JSON or a user specified Go template.`
	_secretLsCommand = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Args:    noSubArgs,
		Short:   "List secrets",
		Long:    secretLsDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			secretLsCommand.InputArgs = args
			secretLsCommand.GlobalFlags = MainGlobalOpts
			secretLsCommand.Remote = remoteclient
			return secretLsCmd(&secretLsCommand)
		},
	}
)

func init() {
	secretLsCommand.Command = _secretLsCommand
	secretLsCommand.SetHelpTemplate(HelpTemplate())
	secretLsCommand.SetUsageTemplate(UsageTemplate())
	flags := secretLsCommand.Flags()

	flags.StringVarP(&secretLsCommand.Filter, "filter", "f", "", "Filter secret output")
	flags.StringVar(&secretLsCommand.Format, "format", "table {{.Driver}}\t{{.Name}}", "Format secret output using Go template")
	flags.BoolVarP(&secretLsCommand.Quiet, "quiet", "q", false, "Print secret output in quiet mode")
}

func secretLsCmd(c *cliconfig.SecretLsValues) error {
	runtime, err := adapter.GetRuntime(getContext(), &c.PodmanCommand)
	if err != nil {
		return errors.Wrapf(err, "error creating libpod runtime")
	}
	defer runtime.Shutdown(false)

	opts := secretLsOptions{
		Quiet: c.Quiet,
	}
	opts.Format = genSecLsFormat(c)

	// Get the filter functions based on any filters set
	var filterFuncs []adapter.SecretFilter
	if c.Filter != "" {
		filters := strings.Split(c.Filter, ",")
		for _, f := range filters {
			filterSplit := strings.Split(f, "=")
			if len(filterSplit) < 2 {
				return errors.Errorf("filter input must be in the form of filter=value: %s is invalid", f)
			}
			generatedFunc, err := generateSecretFilterFuncs(filterSplit[0], filterSplit[1])
			if err != nil {
				return errors.Wrapf(err, "invalid filter")
			}
			filterFuncs = append(filterFuncs, generatedFunc)
		}
	}

	secrets, err := runtime.Secrets(getContext())
	if err != nil {
		return err
	}
	// Get the secrets that match the filter
	secsFiltered := make([]*adapter.Secret, 0, len(secrets))
	for _, sec := range secrets {
		include := true
		for _, filter := range filterFuncs {
			include = include && filter(sec)
		}

		if include {
			secsFiltered = append(secsFiltered, sec)
		}
	}
	return generateSecLsOutput(secsFiltered, opts)
}

// generate the template based on conditions given
func genSecLsFormat(c *cliconfig.SecretLsValues) string {
	var format string
	if c.Format != "" {
		// "\t" from the command line is not being recognized as a tab
		// replacing the string "\t" to a tab character if the user passes in "\t"
		format = strings.Replace(c.Format, `\t`, "\t", -1)
	}
	if c.Quiet {
		format = "{{.Name}}"
	}
	return format
}

// Convert output to genericParams for printing
func secLsToGeneric(templParams []secretLsTemplateParams, JSONParams []secretLsJSONParams) (genericParams []interface{}) {
	if len(templParams) > 0 {
		for _, v := range templParams {
			genericParams = append(genericParams, interface{}(v))
		}
		return
	}
	for _, v := range JSONParams {
		genericParams = append(genericParams, interface{}(v))
	}
	return
}

// generate the accurate header based on template given
func (sec *secretLsTemplateParams) secHeaderMap() map[string]string {
	v := reflect.Indirect(reflect.ValueOf(sec))
	values := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Name
		value := key
		if value == "Name" {
			value = "Secret" + value
		}
		values[key] = strings.ToUpper(splitCamelCase(value))
	}
	return values
}

// getSecTemplateOutput returns all the secrets in the secretLsTemplateParams format
func getSecTemplateOutput(lsParams []secretLsJSONParams, opts secretLsOptions) ([]secretLsTemplateParams, error) {
	var lsOutput []secretLsTemplateParams

	for _, lsParam := range lsParams {
		var (
			labels  string
			options string
		)

		for k, v := range lsParam.Labels {
			label := k
			if v != "" {
				label += "=" + v
			}
			labels += label
		}
		for k, v := range lsParam.Options {
			option := k
			if v != "" {
				option += "=" + v
			}
			options += option
		}
		params := secretLsTemplateParams{
			Name:       lsParam.Name,
			Driver:     lsParam.Driver,
			MountPoint: lsParam.MountPoint,
			Scope:      lsParam.Scope,
			Labels:     labels,
			Options:    options,
		}

		lsOutput = append(lsOutput, params)
	}
	return lsOutput, nil
}

// getSecJSONParams returns the secrets in JSON format
func getSecJSONParams(secrets []*adapter.Secret) []secretLsJSONParams {
	var lsOutput []secretLsJSONParams

	for _, secret := range secrets {
		params := secretLsJSONParams{
			Name:       secret.Name(),
			Labels:     secret.Labels(),
			MountPoint: secret.MountPoint(),
			Driver:     secret.Driver(),
			Options:    secret.Options(),
			Scope:      secret.Scope(),
		}

		lsOutput = append(lsOutput, params)
	}
	return lsOutput
}

// generateSecLsOutput generates the output based on the format, JSON or Go Template, and prints it out
func generateSecLsOutput(secrets []*adapter.Secret, opts secretLsOptions) error {
	if len(secrets) == 0 && opts.Format != formats.JSONString {
		return nil
	}
	lsOutput := getSecJSONParams(secrets)
	var out formats.Writer

	switch opts.Format {
	case formats.JSONString:
		out = formats.JSONStructArray{Output: secLsToGeneric([]secretLsTemplateParams{}, lsOutput)}
	default:
		lsOutput, err := getSecTemplateOutput(lsOutput, opts)
		if err != nil {
			return errors.Wrapf(err, "unable to create secret output")
		}
		out = formats.StdoutTemplateArray{Output: secLsToGeneric(lsOutput, []secretLsJSONParams{}), Template: opts.Format, Fields: lsOutput[0].secHeaderMap()}
	}
	return formats.Writer(out).Out()
}

// generateSecretFilterFuncs returns the true if the secret matches the filter set, otherwise it returns false.
func generateSecretFilterFuncs(filter, filterValue string) (func(secret *adapter.Secret) bool, error) {
	switch filter {
	case "name":
		return func(v *adapter.Secret) bool {
			return strings.Contains(v.Name(), filterValue)
		}, nil
	case "driver":
		return func(v *adapter.Secret) bool {
			return v.Driver() == filterValue
		}, nil
	case "scope":
		return func(v *adapter.Secret) bool {
			return v.Scope() == filterValue
		}, nil
	case "label":
		filterArray := strings.SplitN(filterValue, "=", 2)
		filterKey := filterArray[0]
		if len(filterArray) > 1 {
			filterValue = filterArray[1]
		} else {
			filterValue = ""
		}
		return func(v *adapter.Secret) bool {
			for labelKey, labelValue := range v.Labels() {
				if labelKey == filterKey && ("" == filterValue || labelValue == filterValue) {
					return true
				}
			}
			return false
		}, nil
	case "opt":
		filterArray := strings.SplitN(filterValue, "=", 2)
		filterKey := filterArray[0]
		if len(filterArray) > 1 {
			filterValue = filterArray[1]
		} else {
			filterValue = ""
		}
		return func(v *adapter.Secret) bool {
			for labelKey, labelValue := range v.Options() {
				if labelKey == filterKey && ("" == filterValue || labelValue == filterValue) {
					return true
				}
			}
			return false
		}, nil
	}
	return nil, errors.Errorf("%s is an invalid filter", filter)
}
