package plugin

import "github.com/sonm-io/core/insonmnia/worker/network"

type Config struct {
	SocketDir string        `yaml:"socket_dir" default:"/run/docker/plugins"`
	Volumes   VolumesConfig `yaml:"volume"`
	Overlay   OverlayConfig `yaml:"overlay"`
	GPUs      map[string]map[string]string
}

type VolumesConfig struct {
	Root    string `yaml:"root" default:"/var/lib/docker-volumes"`
	Drivers map[string]map[string]string
}

type OverlayConfig struct {
	Drivers struct {
		Tinc *network.TincNetworkConfig `yaml:"tinc"`
		L2TP *network.L2TPConfig        `yaml:"l2tp"`
	}
}
