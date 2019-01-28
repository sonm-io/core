package network

type NetworkConfig struct {
	// RemoteQOS is an optional QOS server. Used in complex network
	// configurations.
	// You'd likely not want to touch this.
	RemoteQOS string `yaml:"remote_qos"`
}
