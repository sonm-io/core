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
    name: registry.user.dev
    user: name
    password: secret
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)

	// check container description
	assert.Equal(t, "user/image:v1", cfg.GetImageName())
	assert.Equal(t, "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD", cfg.GetSSHKey())

	// volumes
	assert.Contains(t, cfg.Volumes(), "mysmb")
	assert.Contains(t, cfg.Volumes(), "otherType")

	vols := cfg.Volumes()
	assert.Equal(t, "cifs", vols["mysmb"].Type)
	assert.Equal(t, "samba-host.ru/share", vols["mysmb"].Options["share"])
	assert.Equal(t, "username", vols["mysmb"].Options["username"])
	assert.Equal(t, "password", vols["mysmb"].Options["password"])
	assert.Equal(t, "ntlm", vols["mysmb"].Options["security"])
	assert.Equal(t, "3.0", vols["mysmb"].Options["vers"])

	assert.Equal(t, "ipfs", vols["otherType"].Type)

	// mounts
	assert.Contains(t, cfg.Mounts()[0], "mysmb:/mnt:rw")
	assert.Contains(t, cfg.Mounts()[1], "mysmb:/opt:rw")
	assert.Contains(t, cfg.Mounts()[2], "otherType:/home/data:ro")

	// check registry description
	assert.Equal(t, "registry.user.dev", cfg.GetRegistryName())
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
`)
	defer deleteTestConfigFile()

	cfg, err := LoadConfig(testCfgPath)
	assert.NoError(t, err)
	// check explicitly set fileds
	assert.Equal(t, "user/image:v1", cfg.GetImageName())

	// check if non-required fields are empty
	assert.Equal(t, "", cfg.GetSSHKey())
	assert.Equal(t, "", cfg.GetRegistryName())
	assert.Equal(t, "", cfg.GetRegistryAuth())
}

func TestTaskNameRequired(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name:
`)
	defer deleteTestConfigFile()
	cfg, err := LoadConfig(testCfgPath)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestTaskRegistryAuth(t *testing.T) {
	createTestConfigFile(`task:
  container:
    name: user/image:v1
  registry:
    name: registry.user.dev
    user: name
    password: secret
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
