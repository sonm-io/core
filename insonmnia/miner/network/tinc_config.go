package network

type TincNetworkConfig struct {
	Enabled                  bool   `yaml:"enabled"`
	ConfigDir                string `yaml:"config_dir" default:"/tinc"`
	DockerNetPluginSockPath  string `yaml:"docker_net_plugin_dir" default:"/run/docker/plugins/tinc/tinc.sock"`
	DockerIPAMPluginSockPath string `yaml:"docker_ipam_plugin_dir" default:"/run/docker/plugins/tincipam/tincipam.sock"`
	DockerImage              string `yaml:"docker_image" default:"sonm/tinc"`
	StatePath                string `yaml:"state_path" default:"/var/lib/sonm/tinc_network_state"`
}
