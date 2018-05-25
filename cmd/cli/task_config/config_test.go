package task_config

import (
	"io/ioutil"
	"os"
	"testing"

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
        vers: 3.0
    otherType:
      type: ipfs
  mounts:
    - mysmb:/mnt:rw
    - mysmb:/opt:rw
    - otherType:/home/data:ro
registry: 
  username: name
  password: secret
  server_address: registry.user.dev
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
	assert.Equal(t, "3.0", volumes["mysmb"].Options["vers"])

	assert.Equal(t, "ipfs", volumes["otherType"].Type)

	assert.Contains(t, cfg.Container.Mounts[0], "mysmb:/mnt:rw")
	assert.Contains(t, cfg.Container.Mounts[1], "mysmb:/opt:rw")
	assert.Contains(t, cfg.Container.Mounts[2], "otherType:/home/data:ro")

	assert.Equal(t, "name", cfg.Registry.Username)
	assert.Equal(t, "secret", cfg.Registry.Password)
	assert.Equal(t, "registry.user.dev", cfg.Registry.ServerAddress)
}

func TestTaskNoRegistry(t *testing.T) {
	createTestConfigFile(`
container:
  image: user/image:v1
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
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "", cfg.Registry.GetServerAddress())
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
	assert.Equal(t, "", cfg.Registry.GetServerAddress())
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
