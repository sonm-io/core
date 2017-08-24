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

func TestMinerStatusIdle(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerStatus(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.MinerStatusReply{}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Contains(t, out, "  No active tasks\n")
}

func TestMinerStatusData(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerStatus(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&pb.MinerStatusReply{
			Usage: map[string]*pb.ResourceUsage{
				"test": {
					Cpu:    &pb.CPUUsage{Total: uint64(500)},
					Memory: &pb.MemoryUsage{MaxUsage: uint64(2048)},
				},
			},
			Capabilities: &pb.Capabilities{
				Cpu: []*pb.CPUDevice{{Name: "i7", Vendor: "Intel", Mhz: 3000.0, Cores: 4}},
				Gpu: []*pb.GPUDevice{{Name: "GTX 1080Ti", Vendor: "NVidia"}},
				Mem: &pb.RAMDevice{Total: 1000000, Used: 500000},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Contains(t, out, "    CPU0: 4 x i7")
	assert.Contains(t, out, "    GPU0: NVidia GTX 1080Ti")
	assert.Contains(t, out, "      Total: 976.6 KB")
	assert.Contains(t, out, "      Used:  488.3 KB")

	assert.Contains(t, out, "  Tasks")
	assert.Contains(t, out, "    1) test")

	assert.NotContains(t, out, "NET:")
}

func TestMinerStatusJsonIdle(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerStatus(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.MinerStatusReply{}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Equal(t, "{}\n", out)
}

func TestMinerStatusJsonData(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerStatus(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&pb.MinerStatusReply{
			Capabilities: &pb.Capabilities{
				Cpu: []*pb.CPUDevice{{Name: "i7", Vendor: "Intel", Mhz: 3000.0, Cores: 4}},
				Gpu: []*pb.GPUDevice{{Name: "GTX 1080Ti", Vendor: "NVidia"}},
				Mem: &pb.RAMDevice{Total: 1000000, Used: 500000},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	info := &pb.MinerStatusReply{}
	err := json.Unmarshal([]byte(out), &info)
	assert.NoError(t, err)
	assert.NotNil(t, info.Capabilities)

	assert.Equal(t, "Intel", info.Capabilities.Cpu[0].Vendor)
	assert.Equal(t, "i7", info.Capabilities.Cpu[0].Name)
	assert.Equal(t, int32(4), info.Capabilities.Cpu[0].Cores)
	assert.Equal(t, float64(3000), info.Capabilities.Cpu[0].Mhz)

	assert.Equal(t, "NVidia", info.Capabilities.Gpu[0].Vendor)
	assert.Equal(t, "GTX 1080Ti", info.Capabilities.Gpu[0].Name)

	assert.Equal(t, uint64(500000), info.Capabilities.Mem.Used)
	assert.Equal(t, uint64(1000000), info.Capabilities.Mem.Total)
}

func TestMinerStatusFailed(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerStatus(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Equal(t, "[ERR] Cannot get miner status: error\r\n", out)
}

func TestMinerStatusJsonFailed(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerStatus(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, errors.New("some_error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)

	assert.Equal(t, "some_error", cmdErr.Error)
	assert.Equal(t, "Cannot get miner status", cmdErr.Message)
}

func TestMinerListEmpty(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerList(gomock.Any()).AnyTimes().Return(&pb.MinerListReply{}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "No miners connected\r\n", out)
}

func TestMinerListData(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerList(gomock.Any()).
		AnyTimes().
		Return(&pb.MinerListReply{
			Info: map[string]*pb.MinerListReply_ListValue{
				"test": {
					Values: []string{"task-1", "task-2"},
				},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "Miner: test\t\t2 active task(s)\r\n", out)
}

func TestMinerListDataNoTasks(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerList(gomock.Any()).
		AnyTimes().
		Return(&pb.MinerListReply{
			Info: map[string]*pb.MinerListReply_ListValue{
				"test": {},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "Miner: test\t\tIdle\r\n", out)
}

func TestMinerListJsonEmpty(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerList(gomock.Any()).AnyTimes().Return(&pb.MinerListReply{}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()

	assert.Equal(t, "{}\n", out)
}

func TestMinerListJsonData(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerList(gomock.Any()).
		AnyTimes().
		Return(&pb.MinerListReply{
			Info: map[string]*pb.MinerListReply_ListValue{
				"test": {
					Values: []string{"task-1", "task-2"},
				},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeJSON)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()

	reply := &pb.MinerListReply{}
	err := json.Unmarshal([]byte(out), &reply)
	assert.NoError(t, err)

	assert.Len(t, reply.Info, 1)
	minerStat, ok := reply.Info["test"]
	assert.True(t, ok)

	assert.Len(t, minerStat.Values, 2)
	assert.Equal(t, "task-1", minerStat.Values[0])
	assert.Equal(t, "task-2", minerStat.Values[1])
}

func TestMinerListFailed(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerList(gomock.Any()).AnyTimes().Return(nil, errors.New("some_error"))

	buf := initRootCmd(t, config.OutputModeSimple)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()
	assert.Equal(t, "[ERR] Cannot get miners list: some_error\r\n", out)
}

func TestMinerListJsonFailed(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().MinerList(gomock.Any()).AnyTimes().Return(nil, errors.New("some_error"))

	buf := initRootCmd(t, config.OutputModeJSON)
	minerListCmdRunner(rootCmd, itr)
	out := buf.String()

	cmdErr, err := stringToCommandError(out)
	assert.NoError(t, err)
	assert.Equal(t, "some_error", cmdErr.Error)
	assert.Equal(t, "Cannot get miners list", cmdErr.Message)
}

func TestMinerStatusMultiCPUAndGPU(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerStatus(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&pb.MinerStatusReply{
			Capabilities: &pb.Capabilities{
				Cpu: []*pb.CPUDevice{
					{Name: "Xeon E7-4850", Vendor: "Intel", Mhz: 2800.0, Cores: 14},
					{Name: "Xeon E7-8890", Vendor: "Intel", Mhz: 3400.0, Cores: 24},
				},
				Gpu: []*pb.GPUDevice{
					{Name: "GTX 1080Ti", Vendor: "NVidia"},
					{Name: "GTX 1080", Vendor: "NVidia"},
				},
				Mem: &pb.RAMDevice{Total: 1000000, Used: 500000},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Contains(t, out, "CPU0: 14 x Xeon E7-4850")
	assert.Contains(t, out, "CPU1: 24 x Xeon E7-8890")

	assert.Contains(t, out, "GPU0: NVidia GTX 1080Ti")
	assert.Contains(t, out, "GPU1: NVidia GTX 1080")
}

func TestMinerStatusNoGPU(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerStatus(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&pb.MinerStatusReply{
			Capabilities: &pb.Capabilities{
				Cpu: []*pb.CPUDevice{
					{Name: "Xeon E7-4850", Vendor: "Intel", Mhz: 2800.0, Cores: 14},
					{Name: "Xeon E7-8890", Vendor: "Intel", Mhz: 3400.0, Cores: 24},
				},
				Gpu: []*pb.GPUDevice{},
				Mem: &pb.RAMDevice{Total: 1000000, Used: 500000},
			},
		}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Contains(t, out, "GPU: None")
}

func TestMinerStatusWithName(t *testing.T) {
	itr := NewMockCliInteractor(gomock.NewController(t))
	itr.EXPECT().
		MinerStatus(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&pb.MinerStatusReply{
			Name: "fb402dcf-ff56-465e-8aad-bcef7ca1ef9a",
		}, nil)

	buf := initRootCmd(t, config.OutputModeSimple)
	minerStatusCmdRunner(rootCmd, "test", itr)
	out := buf.String()

	assert.Contains(t, out, `Miner: "test" (fb402dcf-ff56-465e-8aad-bcef7ca1ef9a):`)
}
