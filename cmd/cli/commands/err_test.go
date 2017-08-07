package commands

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
