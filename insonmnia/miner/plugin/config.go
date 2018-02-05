package plugin

type Config struct {
	SocketDir string        `yaml:"socket_dir" default:"/run/docker/plugins"`
	Volumes   VolumesConfig `yaml:"volume"`
	GPUs      map[string]map[string]string
}

type VolumesConfig struct {
	Root    string `yaml:"root" default:"/var/lib/docker-volumes"`
	Volumes map[string]map[string]string
}
