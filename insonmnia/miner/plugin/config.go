package plugin

import "github.com/sonm-io/core/insonmnia/miner/network"

type Config struct {
	SocketDir string                     `yaml:"socket_dir" default:"/run/docker/plugins"`
	Volumes   VolumesConfig              `yaml:"volume"`
	Tinc      *network.TincNetworkConfig `yaml:"tinc"`
	L2TP      *network.L2TPConfig        `yaml:"l2tp"`
	GPUs      map[string]map[string]string
}

type VolumesConfig struct {
	Root    string `yaml:"root" default:"/var/lib/docker-volumes"`
	Volumes map[string]map[string]string
}
