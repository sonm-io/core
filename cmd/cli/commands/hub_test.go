package commands

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/cmd/cli/config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestHubPingSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubPing(gomock.Any()).AnyTimes().Return(&pb.PingReply{}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	hubPingCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "OK\n", out)
}

func TestHubPingJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubPing(gomock.Any()).AnyTimes().Return(&pb.PingReply{}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	hubPingCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "{\"status\":\"OK\"}\n", out)
}

func TestHubPingFailed(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubPing(gomock.Any()).AnyTimes().Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	hubPingCmdRunner(rootCmd, itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Ping failed", cmdErr.Message)
}

func TestHubStatus(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubStatus(gomock.Any()).AnyTimes().Return(&pb.HubStatusReply{
		PublicIP:   "1.2.3.4:10002",
		LocalIP:    "10.0.0.1:10002",
		MinerCount: 2,
		Uptime:     1,
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	hubStatusCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "Public Addr:      1.2.3.4:10002\r\nLocal Addr:       10.0.0.1:10002\r\nConnected miners: 2\r\nUptime:           1s\r\n", out)
}

func TestHubStatusJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubStatus(gomock.Any()).AnyTimes().Return(&pb.HubStatusReply{
		PublicIP:   "1.2.3.4:10002",
		LocalIP:    "10.0.0.1:10002",
		MinerCount: 2,
		Uptime:     1,
	}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	hubStatusCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Contains(t, out, `"publicIP":"1.2.3.4:10002"`)
	assert.Contains(t, out, `"localIP":"10.0.0.1:10002"`)
	assert.Contains(t, out, `"minerCount":2`)
	assert.Contains(t, out, `"uptime":1`)
}

func TestHubStatusError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubStatus(gomock.Any()).AnyTimes().Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	hubStatusCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "[ERR] Cannot get status: error\r\n", out)
}

func TestHubStatusJsonError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().HubStatus(gomock.Any()).AnyTimes().Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	hubStatusCmdRunner(rootCmd, itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)

	assert.Equal(t, "Cannot get status", cmdErr.Message)
	assert.Equal(t, "error", cmdErr.Error)
}
