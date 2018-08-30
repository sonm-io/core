package optimus

import (
	"encoding/json"
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
		{Device: &sonm.GPUDevice{ID: "0"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
		{Device: &sonm.GPUDevice{ID: "1"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
		{Device: &sonm.GPUDevice{ID: "2"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
	}
	actual := combinationsGPU(devices, 2)

	expected := [][]*sonm.GPU{
		{
			{Device: &sonm.GPUDevice{ID: "0"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
			{Device: &sonm.GPUDevice{ID: "1"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
		},
		{
			{Device: &sonm.GPUDevice{ID: "0"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
			{Device: &sonm.GPUDevice{ID: "2"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
		},
		{
			{Device: &sonm.GPUDevice{ID: "1"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
			{Device: &sonm.GPUDevice{ID: "2"}, Benchmarks: map[uint64]*sonm.Benchmark{}},
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
	m.EXPECT().SplittingAlgorithm(8).AnyTimes().Return(sonm.SplittingAlgorithm_MIN)
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
			NetFlags:      &sonm.NetFlags{},
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
		0: {ID: 0, Result: 10000},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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
		0: {ID: 0, Result: 10000},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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
		0: {ID: 0, Result: 10000},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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
		3: {ID: 3, Result: 1e9},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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

func TestConsumeCPUAndRAMDoNotStealResourcesWhenRAMFailed(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {ID: 0, Result: 10000},
	}
	devices.RAM.Device.Total = 1e9
	devices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {ID: 3, Result: 1e9},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := sonm.Benchmarks{
		Values: []uint64{5000, 0, 0, 1e10},
	}
	cpuPlan, err := manager.Consume(benchmark, sonm.NetFlags{})
	require.Error(t, err)
	require.Nil(t, cpuPlan)

	// Note that free benchmarks still have to be full, in case of RAM did not fit.
	assert.Equal(t, uint64(10000), manager.freeBenchmarks[0])
	assert.Equal(t, uint64(10000), devices.CPU.Benchmarks[0].Result)
}

func TestConsumeGPU(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			Device: &sonm.GPUDevice{
				Hash: "0",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1200},
				10: {ID: 10, Result: 1860000},
				11: {ID: 11, Result: 3000},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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
			Device: &sonm.GPUDevice{
				Hash: "0",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1200, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 1860000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "1",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1100, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 110000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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
			Device: &sonm.GPUDevice{
				Hash: "0",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1200, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 1860000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "1",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1100, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 110000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 0, 1000, 100000, 3900}
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
			Device: &sonm.GPUDevice{
				Hash: "0",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 100000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "1",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1200, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 120000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "2",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1400, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 140000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "3",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				9:  {ID: 9, Result: 1600, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 160000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 3000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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
		4: {ID: 4, Result: 1e9},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
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

func TestConsumeStorageLowerBound(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.Storage.Device.BytesAvailable = 1e9
	devices.Storage.Benchmarks = map[uint64]*sonm.Benchmark{
		4: {ID: 4, Result: 1e9},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 10}
	plan, err := manager.consumeStorage(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, uint64(64*1<<20), plan.Size.Bytes)

	assert.Equal(t, uint64(1e9-64*1<<20), manager.freeBenchmarks[4])
	assert.Equal(t, uint64(1000e6), devices.Storage.Benchmarks[4].Result)
}

func TestConsumeNetwork(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.Network.In = 100e6
	devices.Network.Out = 100e6
	devices.Network.BenchmarksIn = map[uint64]*sonm.Benchmark{
		5: {ID: 5, Result: 100e6},
	}
	devices.Network.BenchmarksOut = map[uint64]*sonm.Benchmark{
		6: {ID: 6, Result: 100e6},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 5e6, 90e6}
	plan, err := manager.consumeNetwork(benchmark[:], sonm.NetFlags{})
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, uint64(5e6), plan.ThroughputIn.BitsPerSecond)
	assert.Equal(t, uint64(90e6), plan.ThroughputOut.BitsPerSecond)

	assert.Equal(t, uint64(95e6), manager.freeBenchmarks[5])
	assert.Equal(t, uint64(10e6), manager.freeBenchmarks[6])
}

func TestConsumeNetworkWithMultipleIncomingNetFlags(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.Network.NetFlags.SetIncoming(true)
	devices.Network.In = 100e6
	devices.Network.Out = 100e6
	devices.Network.BenchmarksIn = map[uint64]*sonm.Benchmark{
		5: {ID: 5, Result: 100e6},
	}
	devices.Network.BenchmarksOut = map[uint64]*sonm.Benchmark{
		6: {ID: 6, Result: 100e6},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 5e6, 5e6}
	plan, err := manager.consumeNetwork(benchmark[:], sonm.NetFlags{Flags: sonm.NetworkIncoming})
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, uint64(5e6), plan.ThroughputIn.BitsPerSecond)
	assert.Equal(t, uint64(5e6), plan.ThroughputOut.BitsPerSecond)

	assert.Equal(t, uint64(95e6), manager.freeBenchmarks[5])
	assert.Equal(t, uint64(95e6), manager.freeBenchmarks[6])

	plan, err = manager.consumeNetwork(benchmark[:], sonm.NetFlags{Flags: sonm.NetworkIncoming})
	require.Error(t, err)
	require.Nil(t, plan)
}

func TestConsumeRAMMin(t *testing.T) {
	// DEV-718
	devices := newEmptyDevicesReply()
	devices.RAM.Device.Total = 16754622464
	devices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {ID: 3, Result: 16754622464},
	}

	freeDevices := newEmptyDevicesReply()
	freeDevices.RAM.Device.Total = 16754622464
	freeDevices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {ID: 3, Result: 16349238095},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, freeDevices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0}
	ramPlan, err := manager.consumeRAM(benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, ramPlan)

	assert.Equal(t, uint64(sonm.MinRamSize), ramPlan.Size.Bytes)

	assert.Equal(t, uint64(16349238095-sonm.MinRamSize), manager.freeBenchmarks[3])
	assert.Equal(t, uint64(16754622464), devices.RAM.Benchmarks[3].Result)
}

func TestConsumeOrder(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.CPU.Device.Cores = 4
	devices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {ID: 0, Result: 5680},
		1: {ID: 1, Result: 1526},
		2: {ID: 2, Result: 4},
	}
	devices.RAM.Device.Total = 16754622464
	devices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {ID: 3, Result: 16754622464},
	}
	devices.Network.In = 7143572
	devices.Network.Out = 59053206
	devices.Network.BenchmarksIn = map[uint64]*sonm.Benchmark{
		5: {ID: 5, Result: 7143572},
	}
	devices.Network.BenchmarksOut = map[uint64]*sonm.Benchmark{
		6: {ID: 6, Result: 59053206},
	}

	freeDevices := newEmptyDevicesReply()
	freeDevices.CPU.Device.Cores = 4
	freeDevices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {ID: 0, Result: 5183},
		1: {ID: 1, Result: 1526},
		2: {ID: 2, Result: 4},
	}
	freeDevices.RAM.Device.Total = 16754622464
	freeDevices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {ID: 3, Result: 16333650944},
	}
	freeDevices.Network.In = 7143572
	freeDevices.Network.Out = 59053206
	freeDevices.Network.BenchmarksIn = map[uint64]*sonm.Benchmark{
		5: {ID: 5, Result: 6143573},
	}
	freeDevices.Network.BenchmarksOut = map[uint64]*sonm.Benchmark{
		6: {ID: 6, Result: 58053206},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, freeDevices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{80, 80, 1, 4000000, 0, 1, 1}
	{
		plan, err := manager.consumeCPU(benchmark[:])
		require.NoError(t, err)
		require.NotNil(t, plan)

		assert.Equal(t, uint64(6), plan.CorePercents)

		assert.Equal(t, uint64(5183-80), manager.freeBenchmarks[0])
		assert.Equal(t, uint64(5680), devices.CPU.Benchmarks[0].Result)
	}
	{
		plan, err := manager.consumeNetwork(benchmark[:], sonm.NetFlags{})
		require.NoError(t, err)
		require.NotNil(t, plan)

		assert.Equal(t, uint64(1), plan.ThroughputIn.BitsPerSecond)
		assert.Equal(t, uint64(1), plan.ThroughputOut.BitsPerSecond)
	}
}

func TestGPUStrange(t *testing.T) {
	devicesJSON := []byte(`{"CPU":{"device":{"modelName":"Intel(R) Celeron(R) CPU G3900 @ 2.80GHz","cores":2,"sockets":2},"benchmarks":{"0":{"code":"cpu-sysbench-multi","type":1,"image":"sonm/cpu-sysbench@sha256:8eeb78e04954c07b2f72f9311ac2f7eb194456a4af77b2c883f99f8949701924","result":2152,"splittingAlgorithm":1},"1":{"ID":1,"code":"cpu-sysbench-single","type":1,"image":"sonm/cpu-sysbench@sha256:8eeb78e04954c07b2f72f9311ac2f7eb194456a4af77b2c883f99f8949701924","result":1102},"2":{"ID":2,"code":"cpu-cores","type":1,"result":2}}},"GPUs":[{"device":{"ID":"0000:01:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"Memory":3163553792,"hash":"31eeca4424e123f57afd34b054bfcd82"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":277,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":230,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3163553792,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19971000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:02:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"minorNumber":1,"Memory":3165650944,"hash":"f2f2648c5cd8d28a5414d0ee3f0c9f71"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":218,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3165650944,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19934000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:03:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"minorNumber":2,"Memory":3165650944,"hash":"6dc87bc7fd9ff18832bc621a238074a8"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":219,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3165650944,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19956000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:04:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7041,"deviceName":"GeForce GTX 1070","majorNumber":226,"minorNumber":3,"Memory":8513388544,"hash":"e3511750709860779362741ebc4dd762"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":432,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":359,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":8513388544,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":26670000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:05:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"minorNumber":4,"Memory":3165650944,"hash":"3ff9bbd7c4fe4d415762df1a7d6772e0"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":277,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":219,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3165650944,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19977000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:09:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"minorNumber":5,"Memory":3165650944,"hash":"abc271d0428b54d40fc58960c19f639e"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":212,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3165650944,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19945000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:0c:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"minorNumber":6,"Memory":3165650944,"hash":"dd9d1ba7ebca0fdaead5e468458b2efe"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":277,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":213,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3165650944,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19941000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:0d:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7041,"deviceName":"GeForce GTX 1070","majorNumber":226,"minorNumber":7,"Memory":8513388544,"hash":"f67471db2f2ee68525b8d98d1746f74a"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":431,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":338,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":8513388544,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":26628000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:0e:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7044,"deviceName":"GeForce GTX 1060 3GB","majorNumber":226,"minorNumber":8,"Memory":3165650944,"hash":"ad0c41627a7c152eeff224c99b0e7122"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":276,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":213,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":3165650944,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":19937000,"splittingAlgorithm":1}}},{"device":{"ID":"0000:0f:00.0","vendorID":4318,"vendorName":"Nvidia","deviceID":7041,"deviceName":"GeForce GTX 1070","majorNumber":226,"minorNumber":9,"Memory":8513388544,"hash":"9480df065b99e4f2af560f9dee7b5c06"},"benchmarks":{"10":{"ID":10,"code":"gpu-cash-hashrate","type":2,"image":"sonm/gpu-cash-hashrate@sha256:1b4cd1c8a06fcb15762794e50711c5aeab9779d566c75f57d381cee5cde7dfb1","result":430,"splittingAlgorithm":1},"11":{"ID":11,"code":"gpu-redshift","type":2,"image":"sonm/gpu-redshift@sha256:d2b42c0aea94440a01ed90dc71a675868b1dbaf0db012061842571e4983175cc","result":334,"splittingAlgorithm":1},"7":{"ID":7,"code":"gpu-count","type":2,"result":1,"splittingAlgorithm":1},"8":{"ID":8,"code":"gpu-mem","type":2,"result":8513388544,"splittingAlgorithm":2},"9":{"ID":9,"code":"gpu-eth-hashrate","type":2,"image":"sonm/gpu-eth-hashrate@sha256:71ca369ca67b136adcb52df18ce5cb027a1b3e25d63b47aac35c566ec102921e","result":26633000,"splittingAlgorithm":1}}}],"RAM":{"device":{"total":8325287936,"available":8325287936},"benchmarks":{"3":{"ID":3,"code":"ram-size","type":3,"result":8325287936,"splittingAlgorithm":1}}},"network":{"in":11993043,"out":13336076,"netFlags":{"flags":7},"benchmarksIn":{"5":{"ID":5,"code":"net-download","type":5,"image":"sonm/net-bandwidth@sha256:e51c367c5ad56c9ea1dbe1497b4acc7d0839be0832d8b77986b931eedc766fc2","result":11993043,"splittingAlgorithm":1}},"benchmarksOut":{"6":{"ID":6,"code":"net-upload","type":6,"image":"sonm/net-bandwidth@sha256:e51c367c5ad56c9ea1dbe1497b4acc7d0839be0832d8b77986b931eedc766fc2","result":13336076,"splittingAlgorithm":1}}},"storage":{"device":{"bytesAvailable":16030248960},"benchmarks":{"4":{"ID":4,"code":"storage-size","type":4,"result":16030248960,"splittingAlgorithm":1}}}}`)
	devices := newEmptyDevicesReply()
	err := json.Unmarshal(devicesJSON, devices)

	require.NoError(t, err)

	controller := gomock.NewController(t)
	defer controller.Finish()
	{
		manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
		require.NoError(t, err)
		require.NotNil(t, manager)

		benchmark := sonm.Benchmarks{
			Values: []uint64{1000, 800, 1, 1000000, 0, 1000, 1000, 1, 409600000, 84936696, 0, 0},
		}

		plans, err := manager.Consume(benchmark, sonm.NetFlags{})
		require.NoError(t, err)
		require.NotNil(t, plans)

		assert.True(t, len(plans.GPU.Hashes) > 0)
		assert.Equal(t, 4, len(plans.GPU.Hashes))
	}

	{
		manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
		require.NoError(t, err)
		require.NotNil(t, manager)

		benchmark := sonm.Benchmarks{
			Values: []uint64{1000, 800, 1, 1000000, 0, 1000, 1000, 1, 409600000, 218587776, 0, 0},
		}

		plans, err := manager.Consume(benchmark, sonm.NetFlags{})
		require.NoError(t, err)
		require.NotNil(t, plans)

		assert.True(t, len(plans.GPU.Hashes) > 0)
		assert.Equal(t, 10, len(plans.GPU.Hashes))
	}
}

func TestConsumeGPUWithMoreMemoryFails(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			Device: &sonm.GPUDevice{
				Hash: "0",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				8:  {ID: 8, Result: 2.5e9, SplittingAlgorithm: sonm.SplittingAlgorithm_MIN},
				9:  {ID: 9, Result: 1000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "1",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				8:  {ID: 8, Result: 3e9, SplittingAlgorithm: sonm.SplittingAlgorithm_MIN},
				9:  {ID: 9, Result: 1200, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 3.1e9, 2000, 0, 0}
	plan, err := manager.consumeGPU(2, benchmark[:])
	require.Error(t, err)
	require.Nil(t, plan)
}

func TestConsumeGPUWithZeroCountRequiredStillConsumes(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.GPUs = []*sonm.GPU{
		{
			Device: &sonm.GPUDevice{
				Hash: "0",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				8:  {ID: 8, Result: 2.5e9, SplittingAlgorithm: sonm.SplittingAlgorithm_MIN},
				9:  {ID: 9, Result: 1000, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
		{
			Device: &sonm.GPUDevice{
				Hash: "1",
			},
			Benchmarks: map[uint64]*sonm.Benchmark{
				8:  {ID: 8, Result: 3e9, SplittingAlgorithm: sonm.SplittingAlgorithm_MIN},
				9:  {ID: 9, Result: 1200, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				10: {ID: 10, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
				11: {ID: 11, Result: 0, SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL},
			},
		},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 0, 0, 0, 0, 0, 0, 2.5e9, 2000, 0, 0}
	plan, err := manager.consumeGPU(0, benchmark[:])
	require.NoError(t, err)
	require.NotNil(t, plan)

	assert.Equal(t, 2, len(plan.Hashes))
}

func TestConsumeWithCPUCores(t *testing.T) {
	devices := newEmptyDevicesReply()
	devices.CPU.Device.Cores = 2
	devices.CPU.Benchmarks = map[uint64]*sonm.Benchmark{
		0: {ID: 0, Result: 1000},
		1: {ID: 1, Result: 1526},
		2: {ID: 2, Result: 2},
	}
	devices.RAM.Device.Total = 1000000000
	devices.RAM.Benchmarks = map[uint64]*sonm.Benchmark{
		3: {ID: 3, Result: 1000000000},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	manager, err := newDeviceManager(devices, devices, newMappingMock(controller))
	require.NoError(t, err)
	require.NotNil(t, manager)

	benchmark := [12]uint64{0, 0, 4}
	cpuPlan, err := manager.consumeCPU(benchmark[:])
	require.Error(t, err)
	require.Nil(t, cpuPlan)
}
