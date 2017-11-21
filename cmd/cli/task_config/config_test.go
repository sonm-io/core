package task_config

import (
	"io/ioutil"
	"os"
	"testing"

	"encoding/base64"
	"encoding/json"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

const (
	testCfgPath = "test.yaml"
)

func createTestConfigFile(body string) error {
	return ioutil.WriteFile(testCfgPath, []byte(body), 0600)
}

func deleteTestConfigFile() {
	os.Remove(testCfgPath)
}

func TestTaskFull(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
    command: /myapp -param=1
    ssh_key: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD
  registry:
    name: registry.user.dev
    user: name
    password: secret
  resources:
    CPU: 1
    CPU_type: i5
    GPU: true
    GPU_type: nv1080it
    RAM: 10240kb
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)

	// check container description
	assert.Equal(t, "user/image:v1", cfg.GetImageName())
	assert.Equal(t, "/myapp -param=1", cfg.GetEntrypoint())
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD", cfg.GetSSHKey())

	// check registry description
	assert.Equal(t, "registry.user.dev", cfg.GetRegistryName())

	// check resources description
	assert.Equal(t, uint64(1), cfg.GetCPUCount())
	assert.Equal(t, "i5", cfg.GetCPUType())
	assert.Equal(t, true, cfg.GetGPURequirement())
	assert.Equal(t, "nv1080it", cfg.GetGPUType())
	assert.Equal(t, uint64(10485760), cfg.GetRAMCount())
}

func TestTaskNoRegistry(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
    command: /myapp -param=1
    ssh_key: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD
  resources:
    CPU: 1
    CPU_type: i5
    GPU: true
    GPU_type: nv1080it
    RAM: 10240kb
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)

	// check registry description
	assert.Equal(t, "", cfg.GetRegistryName())
	assert.Equal(t, "", cfg.GetRegistryAuth())
}

func TestTaskMinimal(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
  resources:
    RAM: 100MB
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)
	// check explicitly set fileds
	assert.Equal(t, "user/image:v1", cfg.GetImageName())
	assert.Equal(t, uint64(104857600), cfg.GetRAMCount())

	// check if non-required fields are empty
	assert.Equal(t, "", cfg.GetEntrypoint())
	assert.Equal(t, "", cfg.GetSSHKey())
	assert.Equal(t, "", cfg.GetRegistryName())
	assert.Equal(t, "", cfg.GetRegistryAuth())

	// check if defaults are defaults
	assert.Equal(t, uint64(1), cfg.GetCPUCount())
	assert.Equal(t, "any", cfg.GetCPUType())
	assert.Equal(t, false, cfg.GetGPURequirement())
	assert.Equal(t, "any", cfg.GetGPUType())
}

func TestTaskNameRequired(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name:
  resources:
    CPU: 2
    RAM: 100MB
`)
	defer deleteTestConfigFile()
	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestTaskRAMRequired(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
  resources:
    CPU: 1
`)
	defer deleteTestConfigFile()
	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RAM")
	assert.Contains(t, err.Error(), "required")
}

func TestTaskInvaludRAMValue(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
  resources:
    CPU: 1
    RAM: 1488kHz
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse")
	assert.Contains(t, err.Error(), "invalid syntax")
}

func TestTaskRegistryAuth(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
  registry:
    name: registry.user.dev
    user: name
    password: secret
  resources:
    CPU: 1
    RAM: 10MB
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)

	assert.Equal(t, "registry.user.dev", cfg.GetRegistryName())
	auth := cfg.GetRegistryAuth()
	assert.NotEmpty(t, auth)

	authDecoded, err := base64.StdEncoding.DecodeString(auth)
	assert.NoError(t, err)

	authConfig := types.AuthConfig{}
	err = json.Unmarshal(authDecoded, &authConfig)
	assert.NoError(t, err)

	assert.Equal(t, "name", authConfig.Username)
	assert.Equal(t, "secret", authConfig.Password)
	assert.Equal(t, "registry.user.dev", authConfig.ServerAddress)
}

func TestTaskConfigNotExists(t *testing.T) {
	deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)

	assert.Error(t, err)
}

func TestTaskConfigReadError(t *testing.T) {
	body := `task:
  container:
    name: user/image:v1
  resources:
    RAM: 100MB
`
	defer deleteTestConfigFile()

	// write only permissions
	err := ioutil.WriteFile(testCfgPath, []byte(body), 0200)
	assert.NoError(t, err)

	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestTaskConfigEnv(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
    env:
      key1: value1
      key2: value2
  resources:
    RAM: 100MB
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)

	assert.Len(t, cfg.GetEnvVars(), 2)

	env := cfg.GetEnvVars()
	assert.Contains(t, env, "key1")
	assert.Contains(t, env, "key2")

	assert.Equal(t, "value1", env["key1"])
	assert.Equal(t, "value2", env["key2"])
}
