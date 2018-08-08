package task_config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	createTestConfigFile(`
container:
  image: user/image:v1
  ssh_key: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD
  volumes:
    mysmb:
      type: cifs
      options:
        share: samba-host.ru/share
        username: username
        password: password
        security: ntlm
        vers: 3.1
    otherType:
      type: ipfs
  mounts:
    - mysmb:/mnt:rw
    - mysmb:/opt:rw
    - otherType:/home/data:ro
registry: 
  username: name
  password: secret
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "user/image:v1", cfg.Container.Image)
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD", cfg.Container.SshKey)

	assert.Contains(t, cfg.Container.Volumes, "mysmb")
	assert.Contains(t, cfg.Container.Volumes, "otherType")

	volumes := cfg.Container.Volumes
	assert.Equal(t, "cifs", volumes["mysmb"].Type)
	assert.Equal(t, "samba-host.ru/share", volumes["mysmb"].Options["share"])
	assert.Equal(t, "username", volumes["mysmb"].Options["username"])
	assert.Equal(t, "password", volumes["mysmb"].Options["password"])
	assert.Equal(t, "ntlm", volumes["mysmb"].Options["security"])
	assert.Equal(t, "3.1", volumes["mysmb"].Options["vers"])

	assert.Equal(t, "ipfs", volumes["otherType"].Type)

	assert.Contains(t, cfg.Container.Mounts[0], "mysmb:/mnt:rw")
	assert.Contains(t, cfg.Container.Mounts[1], "mysmb:/opt:rw")
	assert.Contains(t, cfg.Container.Mounts[2], "otherType:/home/data:ro")

	assert.Equal(t, "name", cfg.Registry.Username)
	assert.Equal(t, "secret", cfg.Registry.Password)
}

func TestTaskNoRegistry(t *testing.T) {
	createTestConfigFile(`
container:
  image: user/image:v1
  command: /myapp -param=1
  ssh_key: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestTaskMinimal(t *testing.T) {
	createTestConfigFile(`
container:
  image: user/image:v1
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "user/image:v1", cfg.Container.Image)

	assert.Equal(t, "", cfg.Container.SshKey)
}

func TestImageNameRequired(t *testing.T) {
	createTestConfigFile(`
container:
  image:
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image name is required")
}

func TestTaskConfigNotExists(t *testing.T) {
	deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
}

func TestTaskConfigReadError(t *testing.T) {
	body := `
container:
  image: user/image:v1
`
	defer deleteTestConfigFile()

	// Write only permissions.
	err := ioutil.WriteFile(testCfgPath, []byte(body), 0200)
	assert.NoError(t, err)

	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestTaskConfigEnv(t *testing.T) {
	createTestConfigFile(`
container:
  image: user/image:v1
  env:
    key1: value1
    key2: value2
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Len(t, cfg.Container.Env, 2)

	env := cfg.Container.Env
	assert.Contains(t, env, "key1")
	assert.Contains(t, env, "key2")

	assert.Equal(t, "value1", env["key1"])
	assert.Equal(t, "value2", env["key2"])
}

func TestTaskConfigResources(t *testing.T) {
	createTestConfigFile(`
container:
  image: user/image:v1
resources:
  ram:
    size: 30MB
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, uint64(30e6), cfg.GetResources().GetRAM().GetSize().GetBytes())
}

func TestTaskConfigExpose(t *testing.T) {
	createTestConfigFile(`
container:
  image: user/image:v1
  expose:
    - 80:80
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, []string{"80:80"}, cfg.GetContainer().GetExpose())
}

func TestLoadConfigFailsOnUnknownKeys(t *testing.T) {
	createTestConfigFile(`
duration: 240h
price: 0 USD/h

counterparty: 0x6f74d76f4c4b80a61598bded7fca2f660ca742ce
identity: anonymous

resources:
  cpu:
    cores: 0.5
  ram:
    size: 128MB
  storage:
    size: 100MB
  gpu:
    indexes: []
  network:
    throughput_in: 1 Mbit/s  # <- Here bad things
    throughput_out: 1 Mbit/s
    overlay: true
    outbound: true
    incoming: false
`)

	defer deleteTestConfigFile()

	plan := &sonm.AskPlan{}
	err := LoadFromFile(testCfgPath, plan)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "field throughput_in not found")
	assert.Contains(t, err.Error(), "field throughput_out not found")
}
