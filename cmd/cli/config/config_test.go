package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	configDir  = ".sonm"
	configName = "cli.yaml"
)

func getHomeDir() string {
	u, _ := user.Current()
	return u.HomeDir
}

func createTestConfigFile(body string) error {
	dir := path.Join(getHomeDir(), configDir)
	os.Mkdir(dir, 0700)
	cfg := path.Join(dir, configName)
	return ioutil.WriteFile(cfg, []byte(body), 0600)
}

func deleteTestConfigFile() {
	cfg := path.Join(getHomeDir(), configDir, configName)
	os.Remove(cfg)
}

func TestConfigLoad(t *testing.T) {
	err := createTestConfigFile(`
output_format: json
key_store: "/home/user/.sonm/keys/"
pass_phrase: "qwerty123"`)
	defer deleteTestConfigFile()
	assert.NoError(t, err)

	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.Equal(t, "json", cfg.OutputFormat())
	assert.Equal(t, "/home/user/.sonm/keys/", cfg.KeyStore())
	assert.Equal(t, "qwerty123", cfg.PassPhrase())
}

func TestConfigDefaults(t *testing.T) {
	err := createTestConfigFile("")
	defer deleteTestConfigFile()
	assert.NoError(t, err)

	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.OutputFormat())
}

func TestConfigNoFile(t *testing.T) {
	deleteTestConfigFile()

	// no config == all defalts
	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.Equal(t, "simple", cfg.OutputFormat())
}

func TestConfigCannotRead(t *testing.T) {
	defer deleteTestConfigFile()
	dir := path.Join(getHomeDir(), configDir)
	os.Mkdir(dir, 0700)
	cfgPath := path.Join(dir, configName)
	// remove read permissions
	err := ioutil.WriteFile(cfgPath, []byte{}, 0200)
	assert.NoError(t, err)

	cfg, err := NewConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
