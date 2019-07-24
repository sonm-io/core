package sysinit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/docker/docker/pkg/mount"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/action"
	"go.uber.org/zap"
)

const (
	// Path to the "cryptsetup" tool.
	cryptsetupPath = "/sbin/cryptsetup"
)

type Config struct {
	// Device name with Docker partition, for example "/dev/sda2".
	Device string `yaml:"device"`
	// Cipher type.
	Cipher string `yaml:"cipher"`
	// Filesystem type.
	FsType string `yaml:"fs_type"`
	// Where to mount the encrypted Docker partition.
	MountPoint string `yaml:"mount_point"`
}

type InitService struct {
	cfg *Config
	log *zap.SugaredLogger
}

func NewInitService(cfg *Config, log *zap.SugaredLogger) *InitService {
	return &InitService{
		cfg: cfg,
		log: log,
	}
}

func (m *InitService) Reset(ctx context.Context) {
	if err := action.Rollback(m.makeActions()); err != nil {
		m.log.Warnf("failed to reset sysinit service: %v", err)
	}

	m.log.Info("flushed sys/init service")
}

func (m *InitService) Mount(ctx context.Context, request *sonm.InitMountRequest) (*sonm.InitMountResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err, errs := action.NewActionQueue(m.makeActions()...).Execute(ctx)
	if err != nil {
		m.log.Errorw("failed to mount", zap.Error(err))

		if errs != nil {
			m.log.Errorw("failed to rollback mount", zap.Error(errs))
		}

		return nil, err
	}

	return &sonm.InitMountResponse{}, nil
}

func (m *InitService) makeActions() []action.Action {
	name := "td"

	return []action.Action{
		&CreateEncryptedVolumeAction{
			Name:     name,
			Password: "password",
			Device:   m.cfg.Device,
			Cipher:   m.cfg.Cipher,
		},
		&CreateMountPointAction{
			MountPoint: m.cfg.MountPoint,
			Perm:       0755,
		},
		&CreateFileSystemAction{
			Name:   name,
			Type:   m.cfg.FsType,
			Device: m.cfg.Device,
		},
		&MountDeviceMapperAction{
			Name:       name,
			MountPoint: m.cfg.MountPoint,
			Type:       m.cfg.FsType,
			Options:    "",
		},
		&StartDockerAction{},
	}
}

type CreateEncryptedVolumeAction struct {
	Name     string
	Password string
	Device   string
	Cipher   string
}

func (m *CreateEncryptedVolumeAction) Execute(ctx context.Context) error {
	dev := fmt.Sprintf("/dev/%s", m.Device)
	cmd := exec.CommandContext(ctx, cryptsetupPath, "create", m.Name, dev, "--cipher", m.Cipher)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open input pipe for %v: %v", cmd.Args, err)
	}

	if _, err := pipe.Write([]byte(m.Password)); err != nil {
		return fmt.Errorf("failed to write input for %v: %v", cmd.Args, err)
	}

	if err := pipe.Close(); err != nil {
		return fmt.Errorf("failed to close input for %v: %v", cmd.Args, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %v: %v", cmd.Args, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute %v: %v", cmd.Args, err)
	}

	return nil
}

func (m *CreateEncryptedVolumeAction) Rollback() error {
	cmd := exec.Command(cryptsetupPath, "remove", m.Name)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %v: %v", cmd.Args, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute %v: %v", cmd.Args, err)
	}

	return nil
}

type CreateMountPointAction struct {
	MountPoint string
	Perm       os.FileMode
}

func (m *CreateMountPointAction) Execute(ctx context.Context) error {
	if err := os.MkdirAll(m.MountPoint, m.Perm); err != nil {
		return fmt.Errorf("failed to create '%s' mount point: %v", m.MountPoint, err)
	}

	return nil
}

func (m *CreateMountPointAction) Rollback() error {
	if err := os.RemoveAll(m.MountPoint); err != nil {
		return fmt.Errorf("failed to remove '%s' mount point: %v", m.MountPoint, err)
	}

	return nil
}

type CreateFileSystemAction struct {
	Name   string
	Type   string
	Device string
}

func (m *CreateFileSystemAction) Execute(ctx context.Context) error {
	devices, err := ListBlockDevices(ctx)
	if err != nil {
		return err
	}

	device, err := m.findDevice(devices)
	if err != nil {
		return err
	}

	switch len(device.Children) {
	case 0:
		if err := m.makeFileSystem(); err != nil {
			return fmt.Errorf("failed to create filesystem on %s: %v", m.target(), err)
		}

		return nil
	case 1:
		dev := device.Children[0]

		if dev.FsType == m.Type {
			return nil
		}

		if dev.FsType == "" || dev.FsType == "null" {
			if err := m.makeFileSystem(); err != nil {
				return fmt.Errorf("failed to create filesystem on %s: %v", m.target(), err)
			}

			return nil
		}

		return fmt.Errorf("device /dev/%s already has %s filesystem", m.Device, dev.FsType)
	default:
		dev := device.Children[0]
		return fmt.Errorf("device /dev/%s has more than one children: %v", m.Device, dev.Children)
	}
}

func (m *CreateFileSystemAction) Rollback() error {
	// We intentionally do nothing here, i.e. not removing any user data in
	// case possible of errors.
	return nil
}

func (m *CreateFileSystemAction) findDevice(devices []*BlockDevice) (*BlockDevice, error) {
	for _, dev := range devices {
		if dev.Name == m.Device {
			return dev, nil
		}
	}

	return nil, fmt.Errorf("device /dev/%s not found", m.Device)
}

func (m *CreateFileSystemAction) makeFileSystem() error {
	cmd := exec.Command(fmt.Sprintf("mkfs.%s", m.Type), m.target())

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %v: %v", cmd.Args, err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to execute %v: %v", cmd.Args, err)
	}

	return nil
}

func (m *CreateFileSystemAction) target() string {
	return fmt.Sprintf("/dev/mapper/%s", m.Name)
}

type MountDeviceMapperAction struct {
	Name       string
	MountPoint string
	Type       string
	Options    string
}

func (m *MountDeviceMapperAction) Execute(ctx context.Context) error {
	if err := mount.Mount(m.target(), m.MountPoint, m.Type, m.Options); err != nil {
		return fmt.Errorf("failed to mount %s to '%s': %v", m.target(), m.MountPoint, err)
	}

	return nil
}

func (m *MountDeviceMapperAction) Rollback() error {
	if err := mount.Unmount(m.MountPoint); err != nil {
		return fmt.Errorf("failed to unmount %s from '%s': %v", m.target(), m.MountPoint, err)
	}

	return nil
}

func (m *MountDeviceMapperAction) target() string {
	return fmt.Sprintf("/dev/mapper/%s", m.Name)
}

type StartDockerAction struct {
}

func (m *StartDockerAction) Execute(ctx context.Context) error {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := make(chan string)
	if _, err := conn.RestartUnit("docker.service", "fail", ch); err != nil {
		return err
	}

	status := <-ch
	if status != "done" {
		return fmt.Errorf("failed to restart Docker: %s", status)
	}

	return nil
}

func (m *StartDockerAction) Rollback() error {
	conn, err := dbus.New()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch := make(chan string)
	if _, err := conn.StopUnit("docker.service", "fail", ch); err != nil {
		return err
	}

	status := <-ch
	if status != "done" {
		return fmt.Errorf("failed to stop Docker: %s", status)
	}

	return nil
}

type FailAction struct{}

func (m *FailAction) Execute(ctx context.Context) error {
	return fmt.Errorf("%T always fails", m)
}

func (m *FailAction) Rollback() error {
	return fmt.Errorf("%T always fails", m)
}
