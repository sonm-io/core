// +build !darwin,cl

package gpu

import (
	"errors"
	"fmt"
	"log"
	"path"
	"regexp"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sshaman1101/nvidia-docker/nvidia"
)

var (
	// ErrVolumeBadFormat means that volume name does not fit pattern name_version
	ErrVolumeBadFormat = errors.New("bad volume format")
	// ErrVolumeUnsupported means that Volume map does not have such volume
	ErrVolumeUnsupported = errors.New("unsupported volume")
	//ErrVolumeNotFound means that requested volume does not exist
	ErrVolumeNotFound = errors.New("no such volume")
	// ErrVolumeVersion means that version is invalid
	ErrVolumeVersion = errors.New("invalid volume version")

	volVersionRegex = regexp.MustCompile("^([a-zA-Z0-9_.-]+)_([0-9.]+)$")
)

// NewPlugin returns plugin.Driver implementation
// NOTE: volumes is immutable structure!
func NewPlugin(volumes nvidia.VolumeMap) volume.Driver {
	return &volumePlugin{Volumes: volumes}
}

type volumePlugin struct {
	Volumes nvidia.VolumeMap
}

func (p *volumePlugin) Create(req *volume.CreateRequest) error {
	log.Printf("Received create request for volume '%s'", req.Name)
	vol, version, err := p.getVolume(req.Name)
	if err != nil {
		return err
	}

	if version != vol.Version {
		return ErrVolumeVersion
	}

	ok, err := vol.Exists()
	if err != nil {
		return err
	}

	if !ok {
		return vol.Create(nvidia.LinkOrCopyStrategy{})
	}

	return nil
}

func (p *volumePlugin) List() (*volume.ListResponse, error) {
	var resp volume.ListResponse

	for _, vol := range p.Volumes {
		versions, err := vol.ListVersions()
		if err != nil {
			return nil, err
		}

		for _, v := range versions {
			resp.Volumes = append(resp.Volumes, &volume.Volume{
				Name:       fmt.Sprintf("%s_%s", vol.Name, v),
				Mountpoint: path.Join(vol.Path, v),
			})
		}
	}
	return &resp, nil
}

func (p *volumePlugin) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	log.Printf("Received get request for volume '%s'", req.Name)
	vol, version, err := p.getVolume(req.Name)
	if err != nil {
		return nil, err
	}
	ok, err := vol.Exists(version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrVolumeNotFound
	}

	var resp = volume.GetResponse{
		Volume: &volume.Volume{
			Name:       req.Name,
			Mountpoint: path.Join(vol.Path, version),
		},
	}
	return &resp, nil
}

func (p *volumePlugin) Remove(req *volume.RemoveRequest) error {
	log.Printf("Received remove request for volume '%s'", req.Name)
	vol, version, err := p.getVolume(req.Name)
	if err != nil {
		return err
	}

	return vol.Remove(version)
}

func (p *volumePlugin) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	var mReq = volume.MountRequest{
		Name: req.Name,
	}
	mResp, err := p.Mount(&mReq)
	if err != nil {
		return nil, err
	}
	return &volume.PathResponse{Mountpoint: mResp.Mountpoint}, nil
}

func (p *volumePlugin) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	log.Printf("Received mount request for volume '%s'", req.Name)
	vol, version, err := p.getVolume(req.Name)
	if err != nil {
		return nil, err
	}

	ok, err := vol.Exists(version)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrVolumeNotFound
	}

	var resp = volume.MountResponse{
		Mountpoint: path.Join(vol.Path, version),
	}
	return &resp, nil
}

func (p *volumePlugin) Unmount(req *volume.UnmountRequest) error {
	log.Printf("Received unmount request for volume '%s'", req.Name)
	_, _, err := p.getVolume(req.Name)
	return err
}

func (p *volumePlugin) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{Scope: "local"},
	}
}

func (p *volumePlugin) getVolume(name string) (*nvidia.Volume, string, error) {
	m := volVersionRegex.FindStringSubmatch(name)
	if len(m) != 3 {
		return nil, "", ErrVolumeBadFormat
	}

	vol, version := p.Volumes[m[1]], m[2]
	if vol == nil {
		return nil, "", ErrVolumeUnsupported
	}

	return vol, version, nil
}
