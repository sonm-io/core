package sysinit

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFileSystemActionFindDevice(t *testing.T) {
	action := CreateFileSystemAction{
		Name:   "td",
		Type:   "btrfs",
		Device: "sr0",
	}

	devices := []*BlockDevice{
		{
			Name:   "sda",
			FsType: "",
			Children: []*BlockDevice{
				{
					Name:   "sda1",
					FsType: "ntfs",
				},
				{
					Name:   "sda2",
					FsType: "vfat",
				},
				{
					Name: "sda3",
				},
				{
					Name:   "sda4",
					FsType: "ntfs",
				},
				{
					Name:   "sda5",
					FsType: "ext4",
				},
			},
		},
		{
			Name:   "sr0",
			FsType: "",
		},
	}

	dev, err := action.findDevice(devices)

	require.NoError(t, err)
	require.NotNil(t, dev)

	assert.Equal(t, &BlockDevice{
		Name:   "sr0",
		FsType: "",
	}, dev)
}

func TestCreateFileSystemActionFindDeviceChild(t *testing.T) {
	action := CreateFileSystemAction{
		Name:   "td",
		Type:   "btrfs",
		Device: "sda5",
	}

	devices := []*BlockDevice{
		{
			Name:   "sda",
			FsType: "",
			Children: []*BlockDevice{
				{
					Name:   "sda1",
					FsType: "ntfs",
				},
				{
					Name:   "sda2",
					FsType: "vfat",
				},
				{
					Name: "sda3",
				},
				{
					Name:   "sda4",
					FsType: "ntfs",
				},
				{
					Name:   "sda5",
					FsType: "ext4",
				},
			},
		},
		{
			Name:   "sr0",
			FsType: "",
		},
	}

	dev, err := action.findDevice(devices)

	require.NoError(t, err)
	require.NotNil(t, dev)

	assert.Equal(t, &BlockDevice{
		Name:   "sda5",
		FsType: "ext4",
	}, dev)
}
