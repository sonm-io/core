package volume

import (
	"fmt"
	"strings"
)

type Permission uint32

const (
	RW Permission = iota
	RO
)

func ParsePermission(mode string) (Permission, error) {
	switch mode {
	case "rw":
		return RW, nil
	case "ro":
		return RO, nil
	default:
		return RO, fmt.Errorf("invalid permission mode: %s", mode)
	}
}

// NewMount constructs a new mount settings from the Docker representation in
// the following format: `VolumeName:ContainerDestination:ro`.
// For example: `cifs:/mnt:ro`.
func NewMount(spec string) (Mount, error) {
	var mount Mount

	parts, err := parseSpec(spec)
	if err != nil {
		return mount, err
	}

	switch len(parts) {
	case 1:
		mount.Target = parts[0]
	case 2:
		mount.Source = parts[0]
		mount.Target = parts[1]
	case 3:
		permission, err := ParsePermission(parts[2])
		if err != nil {
			return mount, err
		}

		mount.Source = parts[0]
		mount.Target = parts[1]
		mount.Permission = permission
	default:
		return mount, errInvalidSpec(spec)
	}

	return mount, nil
}

// ReadOnly returns true if this mount is read-only.
func (m Mount) ReadOnly() bool {
	return m.Permission == RO
}

// Mount describes mount settings.
type Mount struct {
	Source     string
	Target     string
	Permission Permission
}

func parseSpec(spec string) ([]string, error) {
	if strings.Count(spec, ":") > 2 {
		return nil, errInvalidSpec(spec)
	}

	arr := strings.SplitN(spec, ":", 3)
	if arr[0] == "" {
		return nil, errInvalidSpec(spec)
	}
	return arr, nil
}

func errInvalidSpec(spec string) error {
	return fmt.Errorf("invalid volume specification: %s", spec)
}
