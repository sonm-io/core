package gpu

import (
	"fmt"
	"path"

	"github.com/mitchellh/mapstructure"
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
	name := "nvidia-docker"
	o := &tunerOptions{
		VolumeDriverName: name,
		DriverVersion:    "300.0",
		VolumePath:       fmt.Sprintf("/var/lib/%s/volumes", name),
		SocketPath:       fmt.Sprintf("/run/docker/plugins/%s.sock", name),
	}

	return o
}

func radeonDefaultOptions() *tunerOptions {
	name := "radeon-docker"
	o := &tunerOptions{
		VolumeDriverName: name,
		DriverVersion:    "2482.3",
		VolumePath:       fmt.Sprintf("/var/lib/%s/volumes", name),
		SocketPath:       fmt.Sprintf("/run/docker/plugins/%s.sock", name),
	}

	return o
}

func (opts *tunerOptions) volumeName() string {
	return fmt.Sprintf("%s_%s", opts.VolumeDriverName, opts.DriverVersion)
}
