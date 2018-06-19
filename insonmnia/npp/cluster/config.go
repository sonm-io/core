package cluster

// ClusterConfig represents a cluster membership config.
type Config struct {
	Name      string
	Endpoint  string
	Announce  string
	SecretKey string `yaml:"secret_key" json:"-"`
	Members   []string
}
