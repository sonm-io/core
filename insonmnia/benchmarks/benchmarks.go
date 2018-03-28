package benchmarks

import (
	pb "github.com/sonm-io/core/proto"
)

const (
	// benchmark IDs that must be handled as values from hosts.
	CPUCores    = 3
	RamSize     = 4
	StorageSize = 5
	GPUCount    = 8
	GPUMem      = 9

	BenchIDEnvParamName = "SONM_BENCHMARK_ID"
	CPUCountBenchParam  = "SONM_CPU_COUNT"
)

type BenchList interface {
	List() (map[pb.DeviceType][]*pb.Benchmark, error)
}

type dumbBenchmark struct{}

// NewDumbBenchmarks returns becnhmark list that contains nothing.
func NewDumbBenchmarks() BenchList {
	return &dumbBenchmark{}
}

func (db *dumbBenchmark) List() (map[pb.DeviceType][]*pb.Benchmark, error) {
	return map[pb.DeviceType][]*pb.Benchmark{}, nil
}

// ResultJSON describes results of single benchmark.
type ResultJSON struct {
	Result      uint64 `json:"result"`
	DeviceID    string `json:"device_id"`
	BenchmarkID string `json:"benchmark_id"`
}

// ContainerBenchmarkResultsJSON describes JSON structure which container
// must return as result of one or many benchmarks.
type ContainerBenchmarkResultsJSON struct {
	Results map[uint64]*ResultJSON `json:"results"`
}
