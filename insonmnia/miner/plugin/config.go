package plugin

type Config struct {
	SocketPath string        `yaml:"path"`
	Volumes    VolumesConfig `yaml:"volume"`
}

type VolumesConfig struct {
	Root    string
	Volumes map[string]map[string]string
}
