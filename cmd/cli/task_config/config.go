package task_config

import (
	"fmt"
	"strings"

	b64 "encoding/base64"
	"encoding/json"

	"os"

	ds "github.com/c2h5oh/datasize"
	"github.com/docker/docker/api/types"
	"github.com/jinzhu/configor"
)

// TaskConfig describe how to start task (docker image) on Miner
type TaskConfig interface {
	GetImageName() string
	GetEntrypoint() string
	GetSSHKey() string
	GetEnvVars() map[string]string

	GetRegistryName() string
	GetRegistryAuth() string

	GetMinersPreferred() []string
	GetCPUCount() uint64
	GetRAMCount() uint64
	GetCPUType() string
	GetGPURequirement() bool
	GetGPUType() string
}

type container struct {
	Name       string            `yaml:"name" required:"true"`
	Entrypoint string            `yaml:"command" required:"false"`
	SSHKey     string            `yaml:"ssh_key" required:"false"`
	Env        map[string]string `yaml:"env" required:"false"`
}

type registry struct {
	Name     string `yaml:"name" required:"false"`
	User     string `yaml:"user" required:"false"`
	Password string `yaml:"password" required:"false"`
}

type resources struct {
	CPU     uint64 `yaml:"CPU" required:"true" default:"1"`
	CPUType string `yaml:"CPU_type" required:"false" default:"any"`
	UseGPU  bool   `yaml:"GPU" required:"false" default:"false"`
	GPUType string `yaml:"GPU_type" required:"false" default:"any"`
	RAM     string `yaml:"RAM" required:"true"`
}

type task struct {
	Miners    []string  `yaml:"miners" required:"false"`
	Container container `yaml:"container,flow" required:"true"`
	Resources resources `yaml:"resources,flow" required:"true"`
	Registry  *registry `yaml:"registry,flow" required:"false"`
}

type YamlConfig struct {
	Task task `yaml:"task" required:"true"`
	// RamCount this field is temporary exportable because of bug in configor
	// https://github.com/jinzhu/configor/issues/23
	// maybe it will be better to fork configor and fix this bug
	RamCount uint64 `yaml:"-"`
}

// parseValues check task config internal consistency
func (yc *YamlConfig) parseValues() error {
	var ram ds.ByteSize
	err := ram.UnmarshalText([]byte(strings.ToLower(yc.Task.Resources.RAM)))
	if err != nil {
		return fmt.Errorf("Cannot parse ram: %s", err)
	}
	yc.RamCount = ram.Bytes()
	return nil
}

func (yc *YamlConfig) GetImageName() string {
	return yc.Task.Container.Name
}

func (yc *YamlConfig) GetEntrypoint() string {
	return yc.Task.Container.Entrypoint
}

func (yc *YamlConfig) GetSSHKey() string {
	return yc.Task.Container.SSHKey
}

func (yc *YamlConfig) GetEnvVars() map[string]string {
	return yc.Task.Container.Env
}

func (yc *YamlConfig) GetRegistryName() string {
	if yc.Task.Registry != nil {
		return yc.Task.Registry.Name
	}
	return ""
}

func (yc *YamlConfig) GetRegistryAuth() string {
	if yc.Task.Registry != nil {
		auth := types.AuthConfig{
			Username:      yc.Task.Registry.User,
			Password:      yc.Task.Registry.Password,
			ServerAddress: yc.Task.Registry.Name,
		}
		jsonAuth, _ := json.Marshal(auth)
		return b64.StdEncoding.EncodeToString(jsonAuth)
	}
	return ""
}

func (yc *YamlConfig) GetMinersPreferred() []string {
	return yc.Task.Miners
}

func (yc *YamlConfig) GetCPUCount() uint64 {
	return yc.Task.Resources.CPU
}

func (yc *YamlConfig) GetRAMCount() uint64 {
	return yc.RamCount
}

func (yc *YamlConfig) GetCPUType() string {
	return yc.Task.Resources.CPUType
}

func (yc *YamlConfig) GetGPUType() string {
	return yc.Task.Resources.GPUType
}

func (yc *YamlConfig) GetGPURequirement() bool {
	return yc.Task.Resources.UseGPU
}

func LoadConfig(path string) (TaskConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	conf := &YamlConfig{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}

	err = conf.parseValues()
	if err != nil {
		return nil, err
	}

	return conf, nil
}
