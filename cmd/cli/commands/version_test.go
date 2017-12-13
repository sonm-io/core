package commands

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/stretchr/testify/assert"
)

func initRootCmd(t *testing.T, outFormat string) *bytes.Buffer {
	buf := new(bytes.Buffer)

	cfg := config.NewMockConfig(gomock.NewController(t))
	cfg.EXPECT().OutputFormat().AnyTimes().Return(outFormat)

	Root(cfg)

	rootCmd.ResetCommands()
	rootCmd.ResetFlags()

	rootCmd.SetArgs([]string{""})
	rootCmd.SetOutput(buf)
	return buf
}

func TestGetVersionCmdSimple(t *testing.T) {
	buf := initRootCmd(t, config.OutputModeSimple)

	version = "1.2.3"
	printVersion(rootCmd, version)
	out := buf.String()
	assert.Equal(t, "Version: 1.2.3\r\n", out)
}

func TestGetVersionCmdJson(t *testing.T) {
	buf := initRootCmd(t, config.OutputModeJSON)

	version = "1.2.3"
	printVersion(rootCmd, version)
	out := buf.String()
	assert.Equal(t, "{\"version\":\"1.2.3\"}\r\n", out)
}
