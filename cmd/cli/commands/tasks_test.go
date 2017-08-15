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

	assert.Contains(t, out, "There is 2 tasks on miner \"test\":")
	assert.Contains(t, out, "task-1: RUNNING\r\n")
	assert.Contains(t, out, "task-2: FINISHED\r\n")
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

func TestTaskStartSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStart(gomock.Any(), gomock.Any()).Return(&pb.HubStartTaskReply{
		Id:       "7a94eab1-5f57-485a-8602-124783c588ea",
		Endpoint: []string{"80/tcp->10.0.0.123:12345"},
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStartCmdRunner(rootCmd, "test", "httpd:latest", itr)
	out := buf.String()

	assert.Contains(t, out, "Starting \"httpd:latest\" on miner test...")
	assert.Contains(t, out, "ID 7a94eab1-5f57-485a-8602-124783c588ea, Endpoint [80/tcp->10.0.0.123:12345]")
}

func TestTaskStartJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStart(gomock.Any(), gomock.Any()).Return(&pb.HubStartTaskReply{
		Id:       "7a94eab1-5f57-485a-8602-124783c588ea",
		Endpoint: []string{"80/tcp->10.0.0.123:12345"},
	}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStartCmdRunner(rootCmd, "test", "httpd:latest", itr)
	out := buf.Bytes()

	reply := &pb.HubStartTaskReply{}
	err := json.Unmarshal(out, &reply)
	assert.NoError(t, err)

	assert.Equal(t, "7a94eab1-5f57-485a-8602-124783c588ea", reply.Id)
	assert.Len(t, reply.Endpoint, 1)
	assert.Equal(t, "80/tcp->10.0.0.123:12345", reply.Endpoint[0])
}

func TestTaskStartSimpleError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStart(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStartCmdRunner(rootCmd, "test", "httpd:latest", itr)
	out := buf.String()

	assert.Contains(t, out, "Starting \"httpd:latest\" on miner test...")
	assert.Contains(t, out, "[ERR] Cannot start task: error")
}

func TestTaskStartJsonError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStart(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStartCmdRunner(rootCmd, "test", "httpd:latest", itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Cannot start task", cmdErr.Message)
}

func TestTaskStatusSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status: pb.TaskStatusReply_RUNNING,
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Equal(t, "Task adac72b1-7fcf-47e1-8d74-a53563823185 (on test) status is RUNNING\n", out)
}

func TestTaskStatusJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status: pb.TaskStatusReply_RUNNING,
	}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.Bytes()

	reply := struct {
		ID     string `json:"id"`
		Miner  string `json:"miner"`
		Status string `json:"status"`
	}{}

	err := json.Unmarshal(out, &reply)
	assert.NoError(t, err)
	assert.Equal(t, "adac72b1-7fcf-47e1-8d74-a53563823185", reply.ID)
	assert.Equal(t, "test", reply.Miner)
	assert.Equal(t, pb.TaskStatusReply_RUNNING.String(), reply.Status)
}

func TestTaskStatusSimpleError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Equal(t, "[ERR] Cannot get task status: error\r\n", out)
}

func TestTaskStatusJsonError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Cannot get task status", cmdErr.Message)
}

func TestTaskStopSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStop(gomock.Any(), gomock.Any()).Return(&pb.StopTaskReply{}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStopCmdRunner(rootCmd, "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Equal(t, "OK\n", out)
}

func TestTaskStopJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStop(gomock.Any(), gomock.Any()).Return(&pb.StopTaskReply{}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStopCmdRunner(rootCmd, "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Equal(t, "{\"status\":\"OK\"}\n", out)
}

func TestTaskStopSimpleError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStop(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStopCmdRunner(rootCmd, "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Equal(t, "[ERR] Cannot stop task: error\r\n", out)
}

func TestTaskStopJsonError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStop(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStopCmdRunner(rootCmd, "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Cannot stop task", cmdErr.Message)
}
