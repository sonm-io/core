package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfigDir() string {
	d, _ := ioutil.TempDir("", "sonm_test")
	return d
}

func createTestConfigFile(body string) (string, error) {
	dir := testConfigDir()
	os.Mkdir(dir, 0700)
	cfg := path.Join(dir, configName)
	return dir, ioutil.WriteFile(cfg, []byte(body), 0600)
}

func deleteTestConfigFile(dir string) {
	cfg := path.Join(dir, configName)
	os.Remove(cfg)
}

func TestConfigLoad(t *testing.T) {
	dir, err := createTestConfigFile(`
output_format: json
ethereum:
  key_store: "/home/user/.sonm/keys/"
  pass_phrase: "qwerty123"`)
	defer deleteTestConfigFile(dir)
	assert.NoError(t, err)

	cfg, err := NewConfig(dir)
	assert.NoError(t, err)
	assert.Equal(t, "json", cfg.OutputFormat())
	assert.Equal(t, "/home/user/.sonm/keys/", cfg.KeyStore())
	assert.Equal(t, "qwerty123", cfg.PassPhrase())
}

func TestConfigDefaults(t *testing.T) {
	dir, err := createTestConfigFile("")
	defer deleteTestConfigFile(dir)
	assert.NoError(t, err)

	cfg, err := NewConfig(dir)
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.OutputFormat())
}

func TestConfigNoFile(t *testing.T) {
	// no config == all defalts
	cfg, err := NewConfig(testConfigDir())
	assert.NoError(t, err)
	assert.Equal(t, "simple", cfg.OutputFormat())
}

func TestConfigCannotRead(t *testing.T) {
	dir := testConfigDir()

	os.Mkdir(dir, 0700)
	cfgPath := path.Join(dir, configName)

	defer deleteTestConfigFile(cfgPath)

	// drop read permissions
	err := ioutil.WriteFile(cfgPath, []byte{}, 0200)
	assert.NoError(t, err)

	cfg, err := NewConfig(dir)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestGetConfigPath(t *testing.T) {
	p, err := getConfigPath("/tmp")

	require.NoError(t, err)
	assert.Equal(t, "/tmp/cli.yaml", p)
}
