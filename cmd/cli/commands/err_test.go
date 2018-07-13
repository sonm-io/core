package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/stretchr/testify/assert"
)

func stringToCommandError(s string) (*commandError, error) {
	cmdErr := &commandError{}
	err := json.Unmarshal([]byte(s), &cmdErr)
	if err != nil {
		return nil, err
	}

	return cmdErr, nil
}

func TestJsonErrorInternal(t *testing.T) {
	errStr := newCommandError("test1", fmt.Errorf("err1")).ToJSONString()
	assert.Contains(t, errStr, `"error":"err1"`)
	assert.Contains(t, errStr, `"message":"test1"`)
}

type myError struct {
	code int
	msg  string
}

func (m *myError) Error() string {
	return fmt.Sprintf("%d: %s", m.code, m.msg)
}

func TestJsonErrorCustom(t *testing.T) {
	custom := &myError{code: 123, msg: "some_error"}
	errStr := newCommandError("test2", custom).ToJSONString()
	assert.Contains(t, errStr, `"error":"123: some_error"`)
	assert.Contains(t, errStr, `"message":"test2"`)
}

func TestShowErrorNilErr(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeSimple}
	buf := initRootCmd(t, "1.2.3")
	ShowError(rootCmd, "test error", nil)
	out := buf.String()
	assert.Equal(t, "[ERR] test error\r\n", out)
}

func TestShowErrorWithErr(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeSimple}
	buf := initRootCmd(t, "1.2.3")
	ShowError(rootCmd, "test error", errors.New("internal"))
	out := buf.String()
	assert.Equal(t, "[ERR] test error: internal\r\n", out)
}

func TestShowErrorJsonNilErr(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeJSON}
	buf := initRootCmd(t, "1.2.3")
	ShowError(rootCmd, "test error", nil)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "", cmdErr.Error)
	assert.Equal(t, "test error", cmdErr.Message)
}

func TestShowErrorJsonWithErr(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeJSON}
	buf := initRootCmd(t, "1.2.3")
	ShowError(rootCmd, "test error", errors.New("reason"))
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "reason", cmdErr.Error)
	assert.Equal(t, "test error", cmdErr.Message)
}
