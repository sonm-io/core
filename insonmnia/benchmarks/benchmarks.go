package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
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
	List() map[pb.DeviceType][]*pb.Benchmark
}

type benchmarkList struct {
	data map[pb.DeviceType][]*pb.Benchmark
}

func (bl *benchmarkList) load(ctx context.Context, url string) error {
	ctxlog.G(ctx).Debug("updating benchmarks list")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download benchmarks list: got %s status", resp.Status)
	}

	data := make(map[string]*pb.Benchmark)

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("cannot decode JSON response: %v", err)
	}

	for code, bench := range data {
		key := bench.GetType()
		bench.Code = code

		_, ok := bl.data[key]
		if ok {
			bl.data[key] = append(bl.data[key], bench)
		} else {
			bl.data[key] = []*pb.Benchmark{bench}
		}
	}

	ctxlog.G(ctx).Debug("received benchmarks list", zap.Any("data", bl.data))

	return nil
}

// NewBenchmarksList returns benchmark list from external storage.
func NewBenchmarksList(ctx context.Context, cfg Config) (BenchList, error) {
	ls := &benchmarkList{
		data: make(map[pb.DeviceType][]*pb.Benchmark),
	}

	if err := ls.load(ctx, cfg.URL); err != nil {
		return nil, err
	}

	return ls, nil
}

func (bl *benchmarkList) List() map[pb.DeviceType][]*pb.Benchmark {
	return bl.data
}

// ResultJSON describes results of single benchmark.
type ResultJSON struct {
	Result   uint64 `json:"result"`
	DeviceID string `json:"device_id"`
}

// ContainerBenchmarkResultsJSON describes JSON structure which container
// must return as result of one or many benchmarks.
// Maps benchmark code to result struct
type ContainerBenchmarkResultsJSON struct {
	Results map[string]*ResultJSON `json:"results"`
}

type Config struct {
	URL string `yaml:"url" required:"true"`
}
