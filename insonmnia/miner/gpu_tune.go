package miner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
)

// related to https://github.com/NVIDIA/nvidia-docker/blob/master/src/nvidia-docker-plugin/remote_v1.go#L180
type nvidiaPluginArgs struct {
	VolumeDriver string
	Volumes      []string
	Devices      []string
}

type nvidiaGPUTuner interface {
	Tune(hostconfig *container.HostConfig) error
}

type nilGPUTuner struct{}

func (nilGPUTuner) Tune(hostconfig *container.HostConfig) error {
	return nil
}

type gpuTuner struct {
	args nvidiaPluginArgs
}

func newGPUTuner(config *GPUConfig) (nvidiaGPUTuner, error) {
	// TODO: construct the URL in a more safe way
	resp, err := http.Get("http://" + config.NvidiaDockerDriver + "/docker/cli/json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GPU configuration from nvidia-docker-plugin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b = new(bytes.Buffer)
		io.Copy(b, resp.Body)
		return nil, fmt.Errorf("non OK response from nvidia-docker-plugin: %s %s", resp.Status, b)
	}

	var args nvidiaPluginArgs
	if err = json.NewDecoder(resp.Body).Decode(&args); err != nil {
		return nil, fmt.Errorf("failed to decode GPU args from nvidia-docker-plugin: %v", err)
	}

	return &gpuTuner{args: args}, nil
}

func (g *gpuTuner) Tune(hostconfig *container.HostConfig) error {
	// This tunes configs to get the same result as docker run with:
	// --volume-driver=nvidia-docker --volume=nvidia_driver_375.66:/usr/local/nvidia:ro --device=/dev/nvidiactl --device=/dev/nvidia-uvm --device=/dev/nvidia-uvm-tools --device=/dev/nvidia

	// volumes must be provisioned by docker-nvidia-plugin
	// TODO: can we do the same but w/o plugin? Be a plugin for docker?
	hostconfig.VolumeDriver = g.args.VolumeDriver

	// bind driver volumes inside container
	hostconfig.Binds = g.args.Volumes
	// bind devices inside container
	for _, device := range g.args.Devices {
		hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
			PathOnHost:        device,
			CgroupPermissions: "rwm",
		})
	}

	return nil
}
