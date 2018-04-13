package optimus

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPUCombinationsEmpty(t *testing.T) {
	assert.Equal(t, 0, len(combinationsGPU([]*sonm.GPU{}, 2)))
}

func TestGPUCombinations(t *testing.T) {
	devices := []*sonm.GPU{
		{&sonm.GPUDevice{ID: "0"}, map[uint64]*sonm.Benchmark{}},
		{&sonm.GPUDevice{ID: "1"}, map[uint64]*sonm.Benchmark{}},
		{&sonm.GPUDevice{ID: "2"}, map[uint64]*sonm.Benchmark{}},
	}
	actual := combinationsGPU(devices, 2)

	expected := [][]*sonm.GPU{
		{
			{&sonm.GPUDevice{ID: "0"}, map[uint64]*sonm.Benchmark{}},
			{&sonm.GPUDevice{ID: "1"}, map[uint64]*sonm.Benchmark{}},
		},
		{
			{&sonm.GPUDevice{ID: "0"}, map[uint64]*sonm.Benchmark{}},
			{&sonm.GPUDevice{ID: "2"}, map[uint64]*sonm.Benchmark{}},
		},
		{
			{&sonm.GPUDevice{ID: "1"}, map[uint64]*sonm.Benchmark{}},
			{&sonm.GPUDevice{ID: "2"}, map[uint64]*sonm.Benchmark{}},
		},
	}
	assert.Equal(t, 3, len(actual))
	assert.Equal(t, expected, actual)
}

func newMappingMock(controller *gomock.Controller) *benchmarks.MockMapping {
	m := benchmarks.NewMockMapping(controller)
	m.EXPECT().DeviceType(0).AnyTimes().Return(sonm.DeviceType_DEV_CPU)
	m.EXPECT().DeviceType(1).AnyTimes().Return(sonm.DeviceType_DEV_CPU)
	m.EXPECT().DeviceType(2).AnyTimes().Return(sonm.DeviceType_DEV_CPU)
	m.EXPECT().DeviceType(3).AnyTimes().Return(sonm.DeviceType_DEV_RAM)
	m.EXPECT().DeviceType(4).AnyTimes().Return(sonm.DeviceType_DEV_STORAGE)
	m.EXPECT().DeviceType(5).AnyTimes().Return(sonm.DeviceType_DEV_NETWORK_IN)
	m.EXPECT().DeviceType(6).AnyTimes().Return(sonm.DeviceType_DEV_NETWORK_OUT)
	m.EXPECT().DeviceType(7).AnyTimes().Return(sonm.DeviceType_DEV_GPU)
	m.EXPECT().DeviceType(8).AnyTimes().Return(sonm.DeviceType_DEV_GPU)
	m.EXPECT().DeviceType(9).AnyTimes().Return(sonm.DeviceType_DEV_GPU)
	m.EXPECT().DeviceType(10).AnyTimes().Return(sonm.DeviceType_DEV_GPU)
	m.EXPECT().DeviceType(11).AnyTimes().Return(sonm.DeviceType_DEV_GPU)

	m.EXPECT().SplittingAlgorithm(0).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(1).AnyTimes().Return(sonm.SplittingAlgorithm_NONE)
	m.EXPECT().SplittingAlgorithm(2).AnyTimes().Return(sonm.SplittingAlgorithm_NONE)
	m.EXPECT().SplittingAlgorithm(3).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(4).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(5).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(6).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(7).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(8).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(9).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(10).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)
	m.EXPECT().SplittingAlgorithm(11).AnyTimes().Return(sonm.SplittingAlgorithm_PROPORTIONAL)

	return m
}

func newEmptyDevicesReply() *sonm.DevicesReply {
	return &sonm.DevicesReply{
		CPU: &sonm.CPU{
			Device: &sonm.CPUDevice{
				Cores: 1,
			},
			Benchmarks: map[uint64]*sonm.Benchmark{},
		},
		GPUs: []*sonm.GPU{},
		RAM: &sonm.RAM{
			Device: &sonm.RAMDevice{
				Total: 4e9,
			},
			Benchmarks: map[uint64]*sonm.Benchmark{},
		},
		Network: &sonm.Network{
			BenchmarksIn:  map[uint64]*sonm.Benchmark{},
			BenchmarksOut: map[uint64]*sonm.Benchmark{},
		},
		Storage: &sonm.Storage{
			Device:     &sonm.StorageDevice{},
			Benchmarks: map[uint64]*sonm.Benchmark{},
		},
	}
}

// Note: despite of testing private methods these tests are really helpful.

func TestConsumeCPU(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {
			Result: 10000,
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{5000}
	cpuPlan, err := manager.consumeCPU(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, cpuPlan)

	assert.Equal(t, uint64(50), cpuPlan.CorePercents)

	assert.Equal(t, uint64(5000), manager.freeBenchmarks[0])
	assert.Equal(t, uint64(10000), devices.CPU.Benchmarks[0].Result)
}

func TestConsumeCPUTwoCores(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.CPU.Device.Cores = 2
	devices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {
			Result: 10000,
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{5000}
	cpuPlan, err := manager.consumeCPU(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, cpuPlan)

	assert.Equal(t, uint64(100), cpuPlan.CorePercents)

	assert.Equal(t, uint64(5000), manager.freeBenchmarks[0])
	assert.Equal(t, uint64(10000), devices.CPU.Benchmarks[0].Result)
}

func TestConsumeCPULowerBound(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.CPU.Device.Cores = 2
	devices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {
			Result: 10000,
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0}
	cpuPlan, err := manager.consumeCPU(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, cpuPlan)

	assert.Equal(t, uint64(1), cpuPlan.CorePercents)

	assert.Equal(t, uint64(9950), manager.freeBenchmarks[0])
	assert.Equal(t, uint64(10000), devices.CPU.Benchmarks[0].Result)
}

func TestConsumeRAM(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.RAM.Device.Total = 1e9
	devices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {
			Result: 1e9,
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 700e6}
	ramPlan, err := manager.consumeRAM(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, ramPlan)

	assert.Equal(t, uint64(700e6), ramPlan.Size.Bytes)

	assert.Equal(t, uint64(300e6), manager.freeBenchmarks[3])
	assert.Equal(t, uint64(1000e6), devices.RAM.Benchmarks[3].Result)
}

func TestConsumeGPU(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			&sonm.GPUDevice{
				Hash: "0",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1200},
				10: {Result: 1860000},
				11: {Result: 3000},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 1000, 100000, 2900}
	plan, err := manager.consumeGPU(1, benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, []string{"0"}, plan.Hashes)

	assert.Equal(t, 0, len(manager.freeGPUs))
}

func TestConsumeOneOfTwoGPU(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			&sonm.GPUDevice{
				Hash: "0",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1200},
				10: {Result: 1860000},
				11: {Result: 3000},
			},
		},
		{
			&sonm.GPUDevice{
				Hash: "1",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1100},
				10: {Result: 110000},
				11: {Result: 3000},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 1000, 100000, 2900}
	plan, err := manager.consumeGPU(1, benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, []string{"1"}, plan.Hashes)

	assert.Equal(t, 1, len(manager.freeGPUs))
}

func TestConsumeTwoOfTwoGPU(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			&sonm.GPUDevice{
				Hash: "0",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1200},
				10: {Result: 1860000},
				11: {Result: 3000},
			},
		},
		{
			&sonm.GPUDevice{
				Hash: "1",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1100},
				10: {Result: 110000},
				11: {Result: 3000},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 1000, 100000, 2900}
	plan, err := manager.consumeGPU(2, benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, []string{"0", "1"}, plan.Hashes)

	assert.Equal(t, 0, len(manager.freeGPUs))
}

func TestConsumeTwoOfFourGPU(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			&sonm.GPUDevice{
				Hash: "0",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1000},
				10: {Result: 100000},
				11: {Result: 3000},
			},
		},
		{
			&sonm.GPUDevice{
				Hash: "1",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1200},
				10: {Result: 120000},
				11: {Result: 3000},
			},
		},
		{
			&sonm.GPUDevice{
				Hash: "2",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1400},
				10: {Result: 140000},
				11: {Result: 3000},
			},
		},
		{
			&sonm.GPUDevice{
				Hash: "3",
			},
			map[uint64]*sonm.Benchmark{
				9:  {Result: 1600},
				10: {Result: 160000},
				11: {Result: 3000},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 2300, 200000, 2500}
	plan, err := manager.consumeGPU(2, benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, []string{"0", "2"}, plan.Hashes)

	assert.Equal(t, 2, len(manager.freeGPUs))
}

func TestConsumeStorage(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.Storage.Device.BytesAvailable = 1e9
	devices.Storage.Benchmarks = map[uint64]*sonm.Benchmark{
		4: {
			Result: 1e9,
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 700e6}
	plan, err := manager.consumeStorage(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, uint64(700e6), plan.Size.Bytes)

	assert.Equal(t, uint64(300e6), manager.freeBenchmarks[4])
	assert.Equal(t, uint64(1000e6), devices.Storage.Benchmarks[4].Result)
}

func TestConsumeNetwork(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.Network.In = 100e6
	devices.Network.Out = 100e6
	devices.Network.BenchmarksIn = map[uint64]*sonm.Benchmark{
		5: {Result: 100e6},
	}
	devices.Network.BenchmarksOut = map[uint64]*sonm.Benchmark{
		6: {Result: 100e6},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 5e6, 90e6}
	plan, err := manager.consumeNetwork(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, uint64(5e6), plan.ThroughputIn.BitsPerSecond)
	assert.Equal(t, uint64(90e6), plan.ThroughputOut.BitsPerSecond)

	assert.Equal(t, uint64(95e6), manager.freeBenchmarks[5])
	assert.Equal(t, uint64(10e6), manager.freeBenchmarks[6])
}
