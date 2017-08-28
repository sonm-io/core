package commands

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func getSimpleTaskConfig(t *testing.T, imageName string) task_config.TaskConfig {
	task := task_config.NewMockTaskConfig(gomock.NewController(t))
	task.EXPECT().GetImageName().AnyTimes().Return(imageName)
	task.EXPECT().GetSSHKey().AnyTimes().Return("")
	task.EXPECT().GetRegistryName().AnyTimes().Return("")
	task.EXPECT().GetRegistryAuth().AnyTimes().Return("")

	return task
}

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

	task := getSimpleTaskConfig(t, "httpd:latest")

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStartCmdRunner(rootCmd, "test", task, itr)
	out := buf.String()

	assert.Contains(t, out, "Starting \"httpd:latest\" on miner test...")
	assert.Contains(t, out, "ID 7a94eab1-5f57-485a-8602-124783c588ea")
	assert.Contains(t, out, "Endpoint [80/tcp->10.0.0.123:12345]")
}

func TestTaskStartJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStart(gomock.Any(), gomock.Any()).Return(&pb.HubStartTaskReply{
		Id:       "7a94eab1-5f57-485a-8602-124783c588ea",
		Endpoint: []string{"80/tcp->10.0.0.123:12345"},
	}, nil)

	task := getSimpleTaskConfig(t, "httpd:latest")

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStartCmdRunner(rootCmd, "test", task, itr)
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

	task := getSimpleTaskConfig(t, "httpd:latest")

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStartCmdRunner(rootCmd, "test", task, itr)
	out := buf.String()

	assert.Contains(t, out, "Starting \"httpd:latest\" on miner test...")
	assert.Contains(t, out, "[ERR] Cannot start task: error")
}

func TestTaskStartJsonError(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStart(gomock.Any(), gomock.Any()).Return(nil, errors.New("error"))

	task := getSimpleTaskConfig(t, "httpd:latest")

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStartCmdRunner(rootCmd, "test", task, itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "error", cmdErr.Error)
	assert.Equal(t, "Cannot start task", cmdErr.Message)
}

func TestTaskStatusSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status:    pb.TaskStatusReply_RUNNING,
		ImageName: "httpd:latest",
		Uptime:    60,
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Equal(t, "Task adac72b1-7fcf-47e1-8d74-a53563823185 (on test):\r\n  Image:  httpd:latest\r\n  Status: RUNNING\r\n  Uptime: 60ns\r\n", out)
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

func TestTaskStatusWithPortsSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status:    pb.TaskStatusReply_RUNNING,
		ImageName: "httpd:latest",
		Ports:     `{"80/tcp":[{"HostIp":"0.0.0.0","HostPort":"32775"}],"8080/tcp":[{"HostIp":"0.0.0.0","HostPort":"32777"}]}`,
		Uptime:    60,
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Contains(t, out, "  Ports:\r\n")
	assert.Contains(t, out, "    80/tcp: 0.0.0.0:32775\r\n")
	assert.Contains(t, out, "    8080/tcp: 0.0.0.0:32777\r\n")
}

func TestTaskStatusWithInvalidPortsSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status:    pb.TaskStatusReply_RUNNING,
		ImageName: "httpd:latest",
		Ports:     `{"invalid": "input"}`,
		Uptime:    60,
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.NotContains(t, out, "  Ports:\r\n")
}

func TestTaskStatusWithPortsJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status:    pb.TaskStatusReply_RUNNING,
		ImageName: "httpd:latest",
		Ports:     `{"80/tcp":[{"HostIp":"0.0.0.0","HostPort":"32775"}],"8080/tcp":[{"HostIp":"0.0.0.0","HostPort":"32777"}]}`,
		Uptime:    60,
	}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)

	outJson := map[string]string{}
	err := json.Unmarshal(buf.Bytes(), &outJson)
	assert.NoError(t, err)

	assert.Len(t, outJson, 6, "Must have 6 fields")

	assert.Contains(t, outJson, "id")
	assert.Equal(t, "adac72b1-7fcf-47e1-8d74-a53563823185", outJson["id"])

	assert.Contains(t, outJson, "miner")
	assert.Equal(t, "test", outJson["miner"])

	assert.Contains(t, outJson, "status")
	assert.Equal(t, "RUNNING", outJson["status"])

	assert.Contains(t, outJson, "image")
	assert.Equal(t, "httpd:latest", outJson["image"])

	assert.Contains(t, outJson, "uptime")
	assert.Equal(t, "60", outJson["uptime"])

	// extra checks for ports
	assert.Contains(t, outJson, "ports")
	ports := nat.PortMap{}
	err = json.Unmarshal([]byte(outJson["ports"]), &ports)
	assert.NoError(t, err)
	assert.Len(t, ports, 2)

	binding1, ok := ports["80/tcp"]
	assert.True(t, ok)
	assert.Len(t, binding1, 1)
	assert.Equal(t, "0.0.0.0", binding1[0].HostIP)
	assert.Equal(t, "32775", binding1[0].HostPort)

	binding2, ok := ports["8080/tcp"]
	assert.True(t, ok)
	assert.Len(t, binding2, 1)
	assert.Equal(t, "0.0.0.0", binding2[0].HostIP)
	assert.Equal(t, "32777", binding2[0].HostPort)
}

func TestTaskStatusWithResourcesSimple(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status:    pb.TaskStatusReply_RUNNING,
		ImageName: "httpd:latest",
		Usage: &pb.ResourceUsage{
			Cpu:    &pb.CPUUsage{Total: 10000},
			Memory: &pb.MemoryUsage{MaxUsage: 104857600}, // 100 mb
			Network: map[string]*pb.NetworkUsage{
				"eth0": {
					TxBytes:   1,
					RxBytes:   2,
					TxPackets: 3,
					RxPackets: 4,
					TxErrors:  5,
					RxErrors:  6,
					TxDropped: 7,
					RxDropped: 8,
				},
			},
		},
	}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)
	out := buf.String()

	assert.Contains(t, out, "  Resources:\n")
	assert.Contains(t, out, "    CPU: 10000\r\n")
	assert.Contains(t, out, "    MEM: 100.0 MB\r\n")

	assert.Contains(t, out, "    NET:")
	assert.Contains(t, out, "      eth0:")
	assert.Contains(t, out, "        Tx/Rx bytes: 1/2")
	assert.Contains(t, out, "        Tx/Rx packets: 3/4")
	assert.Contains(t, out, "        Tx/Rx errors: 5/6")
	assert.Contains(t, out, "        Tx/Rx dropped: 7/8")
}

func TestTaskStatusWithResourcesJson(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().TaskStatus(gomock.Any(), gomock.Any()).Return(&pb.TaskStatusReply{
		Status:    pb.TaskStatusReply_RUNNING,
		ImageName: "httpd:latest",
		Ports:     `{"80/tcp":[{"HostIp":"0.0.0.0","HostPort":"32775"}],"8080/tcp":[{"HostIp":"0.0.0.0","HostPort":"32777"}]}`,
		Uptime:    60,
		Usage: &pb.ResourceUsage{
			Cpu:    &pb.CPUUsage{Total: 10000},
			Memory: &pb.MemoryUsage{MaxUsage: 104857600}, // 100 mb
			Network: map[string]*pb.NetworkUsage{
				"eth0": {
					TxBytes:   1,
					RxBytes:   2,
					TxPackets: 3,
					RxPackets: 4,
					TxErrors:  5,
					RxErrors:  6,
					TxDropped: 7,
					RxDropped: 8,
				},
			},
		},
	}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	taskStatusCmdRunner(rootCmd, "test", "adac72b1-7fcf-47e1-8d74-a53563823185", itr)

	outJson := map[string]interface{}{}
	err := json.Unmarshal(buf.Bytes(), &outJson)
	assert.NoError(t, err)

	assert.Len(t, outJson, 9, "Must have 8 fields")

	assert.Contains(t, outJson, "cpu")
	assert.Equal(t, "10000", outJson["cpu"])
	assert.Contains(t, outJson, "mem")
	assert.Equal(t, "104857600", outJson["mem"])
	assert.Contains(t, outJson, "net")

	net := outJson["net"].(map[string]interface{})
	assert.Contains(t, net, "eth0")

	eth0 := net["eth0"].(map[string]interface{})
	assert.Contains(t, eth0, "txBytes")
	assert.Equal(t, float64(1), eth0["txBytes"].(float64))

	assert.Contains(t, eth0, "rxBytes")
	assert.Equal(t, float64(2), eth0["rxBytes"].(float64))

	assert.Contains(t, eth0, "txPackets")
	assert.Equal(t, float64(3), eth0["txPackets"].(float64))

	assert.Contains(t, eth0, "rxPackets")
	assert.Equal(t, float64(4), eth0["rxPackets"].(float64))

	assert.Contains(t, eth0, "txErrors")
	assert.Equal(t, float64(5), eth0["txErrors"].(float64))

	assert.Contains(t, eth0, "rxErrors")
	assert.Equal(t, float64(6), eth0["rxErrors"].(float64))

	assert.Contains(t, eth0, "txDropped")
	assert.Equal(t, float64(7), eth0["txDropped"].(float64))

	assert.Contains(t, eth0, "rxDropped")
	assert.Equal(t, float64(8), eth0["rxDropped"].(float64))
}
