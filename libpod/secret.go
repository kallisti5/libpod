package libpod

// Secret is the type used to create named secrets
// TODO: all secrets should be created using this and the Secret API
type Secret struct {
	config *SecretConfig

	valid   bool
	runtime *Runtime
}

// SecretConfig holds the secret's config information
//easyjson:json
type SecretConfig struct {
	// Name of the secret
	Name string `json:"name"`

	Labels        map[string]string `json:"labels"`
	Driver        string            `json:"driver"`
	Scope         string            `json:"scope"`
	IsCtrSpecific bool              `json:"ctrSpecific"`
}

// Name retrieves the secret's name
func (v *Secret) Name() string {
	return v.config.Name
}

// Labels returns the secret's labels
func (v *Secret) Labels() map[string]string {
	labels := make(map[string]string)
	for key, value := range v.config.Labels {
		labels[key] = value
	}
	return labels
}

// MountPoint returns the secret's mountpoint on the host
func (v *Secret) MountPoint() string {
	return "/secrets/" + v.config.Driver + "/" + v.config.Name
}

// Driver returns the secret's driver
func (v *Secret) Driver() string {
	return v.config.Driver
}

// Scope returns the scope of the secret
func (v *Secret) Scope() string {
	return v.config.Scope
}

// IsCtrSpecific returns whether this secret was created specifically for a
// given container. Images with this set to true will be removed when the
// container is removed with the Secrets parameter set to true.
func (v *Secret) IsCtrSpecific() bool {
	return v.config.IsCtrSpecific
}
