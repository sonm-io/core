package gpu

import (
	"fmt"
	"path"

	"github.com/mitchellh/mapstructure"
)

const (
	nvidiaVolumeDriver  = "nvidia-docker"
	nvidiaDriverVersion = "300.0"
	radeonVolumeDriver  = "radeon-docker"
	radeonDriverVersion = "2482.3"
)

// tunerOptions contains various options for embedded GPU tuners
type tunerOptions struct {
	VolumeDriverName string `mapstructure:"driver_name"`
	DriverVersion    string `mapstructure:"driver_version"`
	VolumePath       string `mapstructure:"volume_path"`
	SocketPath       string `mapstructure:"-"`
}

type Option func(*tunerOptions)

func WithSocketDir(dir string) Option {
	return func(opts *tunerOptions) {
		sock := opts.VolumeDriverName + ".sock"
		opts.SocketPath = path.Join(dir, sock)
	}
}

func WithOptions(raw map[string]string) Option {
	return func(opts *tunerOptions) {
		mapstructure.Decode(raw, &opts)
	}
}

func nvidiaDefaultOptions() *tunerOptions {
	return &tunerOptions{
		VolumeDriverName: nvidiaVolumeDriver,
		DriverVersion:    nvidiaDriverVersion,
		VolumePath:       fmt.Sprintf("/var/lib/%s/volumes", nvidiaVolumeDriver),
		SocketPath:       fmt.Sprintf("/run/docker/plugins/%s.sock", nvidiaVolumeDriver),
	}
}

func radeonDefaultOptions() *tunerOptions {
	return &tunerOptions{
		VolumeDriverName: radeonVolumeDriver,
		DriverVersion:    radeonDriverVersion,
		VolumePath:       fmt.Sprintf("/var/lib/%s/volumes", radeonVolumeDriver),
		SocketPath:       fmt.Sprintf("/run/docker/plugins/%s.sock", radeonVolumeDriver),
	}
}

func (opts *tunerOptions) volumeName() string {
	return fmt.Sprintf("%s_%s", opts.VolumeDriverName, opts.DriverVersion)
}
