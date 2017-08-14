package commands

import (
	"testing"

	"errors"

	"encoding/json"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/cmd/cli/config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestHubPingSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubPing().AnyTimes().Return(&pb.PingReply{}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	hubPingCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "OK\n", out)
}

func TestHubPingJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubPing().AnyTimes().Return(&pb.PingReply{}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	hubPingCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "{\"status\":\"OK\"}\n", out)
}

func TestHubPingFailed(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubPing().AnyTimes().Return(nil, errors.New("error"))

	buf := initRootCmd(t, "1")
	hubPingCmdRunner(rootCmd, itr)
	out := buf.String()

	cmdErr := &commandError{}
	err := json.Unmarshal([]byte(out), &cmdErr)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Ping failed", cmdErr.Message)
}
