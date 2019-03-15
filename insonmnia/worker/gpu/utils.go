package gpu

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/sonm-io/core/proto"
)

func newVolumeMount(src, dst, name string) mount.Mount {
	return mount.Mount{
		Type:         mount.TypeVolume,
		Source:       src,
		Target:       dst,
		ReadOnly:     true,
		Consistency:  mount.ConsistencyDefault,
		BindOptions:  nil,
		TmpfsOptions: nil,
		VolumeOptions: &mount.VolumeOptions{
			NoCopy: false,
			Labels: map[string]string{},
			DriverConfig: &mount.Driver{
				Name:    name,
				Options: map[string]string{},
			},
		},
	}
}

func tuneContainer(hostconfig *container.HostConfig, devices map[GPUID]*sonm.GPUDevice, ids []GPUID) error {
	var cardsToBind = make(map[GPUID]*sonm.GPUDevice)
	var volumeMapping = make(map[string]mount.Mount)
	for _, id := range ids {
		card, ok := devices[id]
		if !ok {
			return fmt.Errorf("cannot allocate device: unknown id %s", id)
		}

		cardsToBind[id] = card
	}

	for _, card := range cardsToBind {
		for _, device := range card.GetDeviceFiles() {
			hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
				PathOnHost:        device,
				PathInContainer:   device,
				CgroupPermissions: "rwm",
			})
		}

		for name, pair := range card.GetDriverVolumes() {
			srcDst := strings.Split(pair, ":")
			if len(srcDst) != 2 {
				return fmt.Errorf("malformed driver mount-point `%s`", pair)
			}

			volumeMapping[srcDst[1]] = newVolumeMount(srcDst[0], srcDst[1], name)
		}
	}

	for _, mnt := range volumeMapping {
		hostconfig.Mounts = append(hostconfig.Mounts, mnt)
	}

	return nil
}
