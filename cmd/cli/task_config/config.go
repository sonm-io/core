package task_config

import (
	b64 "encoding/base64"
	"encoding/json"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/proto"
)

// TaskConfig describe how to start task (docker image) on Miner
type TaskConfig interface {
	GetImageName() string
	GetSSHKey() string
	GetEnvVars() map[string]string
	GetCommitOnStop() bool

	GetRegistryName() string
	GetRegistryAuth() string

	Volumes() map[string]volume
	Mounts() []string
	Networks() []network
	GetResources() *sonm.AskPlanResources
}

type container struct {
	Name         string            `yaml:"name" required:"true"`
	Entrypoint   string            `yaml:"command" required:"false"`
	SSHKey       string            `yaml:"ssh_key" required:"false"`
	Env          map[string]string `yaml:"env" required:"false"`
	CommitOnStop bool              `yaml:"commit_on_stop" required:"false"`
	Volumes      map[string]volume
	Mounts       []string
	Networks     []network
}

type volume struct {
	Type    string            `yaml:"type" required:"true"`
	Options map[string]string `yaml:"options" required:"false"`
}

type network struct {
	Type    string            `yaml:"type,omitempty"`
	Options map[string]string `yaml:"options,omitempty"`
	Subnet  string            `yaml:"subnet,omitempty"`
	Addr    string            `yaml:"addr,omitempty"`
}

type registry struct {
	Name     string `yaml:"name" required:"false"`
	User     string `yaml:"user" required:"false"`
	Password string `yaml:"password" required:"false"`
}

type task struct {
	Container container              `yaml:"container,flow" required:"true"`
	Registry  *registry              `yaml:"registry,flow" required:"false"`
	Resources *sonm.AskPlanResources `yaml:"resources,flow"`
}

type YamlConfig struct {
	Task task `yaml:"task" required:"true"`
}

func (yc *YamlConfig) GetImageName() string {
	return yc.Task.Container.Name
}

func (yc *YamlConfig) GetSSHKey() string {
	return yc.Task.Container.SSHKey
}

func (yc *YamlConfig) GetEnvVars() map[string]string {
	return yc.Task.Container.Env
}

func (yc *YamlConfig) GetCommitOnStop() bool {
	return yc.Task.Container.CommitOnStop
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

func (yc *YamlConfig) Volumes() map[string]volume {
	return yc.Task.Container.Volumes
}

func (yc *YamlConfig) Mounts() []string {
	return yc.Task.Container.Mounts
}

func (yc *YamlConfig) Networks() []network {
	return yc.Task.Container.Networks
}

func (yc *YamlConfig) GetResources() *sonm.AskPlanResources {
	return yc.Task.Resources
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

	return conf, nil
}
