package plugin

import "github.com/sonm-io/core/insonmnia/miner/network"

type Config struct {
	SocketDir string        `yaml:"socket_dir" default:"/run/docker/plugins"`
	Volumes   VolumesConfig `yaml:"volume"`
	GPUs      map[string]map[string]string
	Tinc      *network.TincNetworkConfig `yaml:"tinc"`
}

type VolumesConfig struct {
	Root    string `yaml:"root" default:"/var/lib/docker-volumes"`
	Volumes map[string]map[string]string
}
