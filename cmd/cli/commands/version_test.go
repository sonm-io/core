package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/stretchr/testify/assert"
)

func initRootCmd(t *testing.T, ver string) *bytes.Buffer {
	buf := new(bytes.Buffer)

	Root(ver)

	rootCmd.ResetCommands()
	rootCmd.ResetFlags()

	rootCmd.SetArgs([]string{""})
	rootCmd.SetOutput(buf)
	return buf
}

func TestGetVersionCmdSimple(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeSimple}
	buf := initRootCmd(t, "1.2.3")

	printVersion(rootCmd, version)
	out := buf.String()
	assert.Contains(t, out, "sonmcli 1.2.3")
}

func TestGetVersionCmdJson(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeJSON}
	buf := initRootCmd(t, "1.2.3")

	printVersion(rootCmd, version)

	v := make(map[string]string)
	err := json.Unmarshal(buf.Bytes(), &v)
	assert.NoError(t, err)

	assert.Contains(t, v, "version")
	assert.Contains(t, v, "platform")
}
