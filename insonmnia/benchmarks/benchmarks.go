package benchmarks

import (
	pb "github.com/sonm-io/core/proto"
)

const (
	// we can get values for this benchmarks from the host system
	CPUCores    = "cpu-cores"
	RamSize     = "ram-size"
	StorageSize = "storage-size"
	GPUCount    = "gpu-count"
	GPUMem      = "gpu-mem"

	// this benchmarks executed into a container
	CPUSysbenchSingle = "cpu-sysbench-single"
	CPUSysbenchMulti  = "cpu-sysbench-multi"
	NetDownload       = "net-download"
	NetUpload         = "net-upload"
	GPUEthHashrate    = "gpu-eth-hashrate"
	GPUCashHashrate   = "gpu-cash-hashrate"
	GPURedshift       = "gpu-redshift"
)

type BenchList interface {
	List() map[pb.DeviceType][]*pb.Benchmark
}

type dumbBenchmark struct{}

// NewDumbBenchmarks returns becnhmark list that contains nothing.
func NewDumbBenchmarks() BenchList {
	return &dumbBenchmark{}
}

func (db *dumbBenchmark) List() map[pb.DeviceType][]*pb.Benchmark {
	return map[pb.DeviceType][]*pb.Benchmark{}
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
	Results map[string]*ResultJSON `json:"results"`
}
