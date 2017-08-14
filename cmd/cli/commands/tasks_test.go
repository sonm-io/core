package commands

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/cmd/cli/config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestTasksListSimpleEmpty(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(&pb.StatusMapReply{}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskListCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Equal(t, "There is no tasks on miner \"test\"\r\n", out)
}

func TestTasksListSimpleWithTasks(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(&pb.StatusMapReply{
		Statuses: map[string]*pb.TaskStatusReply{
			"task-1": {Status: pb.TaskStatusReply_RUNNING},
			"task-2": {Status: pb.TaskStatusReply_FINISHED},
		},
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskListCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Equal(t, "There is 2 tasks on miner \"test\":\r\n  task-1: RUNNING\r\n  task-2: FINISHED\r\n", out)
}

func TestTaskListJsonEmpty(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(&pb.StatusMapReply{}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskListCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Equal(t, "{}\n", out)
}

func TestTaskListJsonWithTasks(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(&pb.StatusMapReply{
		Statuses: map[string]*pb.TaskStatusReply{
			"task-1": {Status: pb.TaskStatusReply_RUNNING},
			"task-2": {Status: pb.TaskStatusReply_FINISHED},
		},
	}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskListCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	reply := &pb.StatusMapReply{}
	err := json.Unmarshal([]byte(out), &reply)
	assert.NoError(t, err)
	assert.Len(t, reply.Statuses, 2)

	status, ok := reply.Statuses["task-1"]
	assert.True(t, ok)
	assert.Equal(t, pb.TaskStatusReply_RUNNING, status.GetStatus())
}

func TestTaskListSimpleError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	taskListCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Equal(t, "[ERR] Cannot get tasks: error\r\n", out)
}

func TestTaskListJsonError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskList(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	taskListCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Cannot get tasks", cmdErr.Message)
}
