// +build !darwin,cl

package gpu

import (
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/sonm-io/core/proto"
)

type radeonTuner struct {
	volumePluginHandler
}

func newRadeonTuner(opts ...Option) (Tuner, error) {
	options := radeonDefaultOptions()
	for _, f := range opts {
		f(options)
	}

	tun := radeonTuner{}
	tun.options = options

	if err := hasGPUWithVendor(sonm.GPUVendorType_RADEON); err != nil {
		return nil, err
	}

	tun.devices = tun.getDevices()
	return tun, nil
}

func (radeonTuner) getDevices() []string {
	var dev []string
	if _, err := os.Stat("/dev/dri"); err == nil {
		dev = append(dev, "/dev/dri")
	}

	if _, err := os.Stat("/dev/kfd"); err == nil {
		dev = append(dev, "/dev/kfd")
	}

	return dev
}

func (tun radeonTuner) Tune(hostconfig *container.HostConfig) error {
	return tun.tune(hostconfig)
}

func (tun radeonTuner) Close() error {
	if err := tun.listener.Close(); err != nil {
		return err
	}

	return os.Remove(tun.options.SocketPath)
}
