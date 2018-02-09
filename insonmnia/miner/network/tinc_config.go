package network

type TincNetworkConfig struct {
	Enabled                  bool   `yaml:"enabled"`
	ConfigDir                string `yaml:"config_dir" default:"/tmp/tinc"`
	DockerNetPluginSockPath  string `yaml:"docker_net_plugin_dir" default:"/run/docker/plugins/tinc/tinc.sock"`
	DockerIPAMPluginSockPath string `yaml:"docker_ipam_plugin_dir" default:"/run/docker/plugins/tincipam/tincipam.sock"`
}
