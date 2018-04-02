package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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

func (bl *benchmarkList) load(ctx context.Context, s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse input as URL: %v", err)
	}

	var reader io.ReadCloser
	switch u.Scheme {
	case "http", "https":
		reader, err = bl.loadURL(ctx, u.String())
	case "file":
		reader, err = bl.loadFile(ctx, u.Path)
	default:
		err = fmt.Errorf("unknown url schema for downloading: %v", u.Scheme)
	}

	if err != nil {
		return err
	}

	defer reader.Close()
	return bl.readResults(ctx, reader)
}

func (bl *benchmarkList) loadURL(ctx context.Context, url string) (io.ReadCloser, error) {
	ctxlog.G(ctx).Debug("loading benchmark list url", zap.String("url", url))

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download benchmarks list: got %s status", resp.Status)
	}

	return resp.Body, nil
}

func (bl *benchmarkList) loadFile(ctx context.Context, path string) (io.ReadCloser, error) {
	ctxlog.G(ctx).Debug("loading benchmark list from file", zap.String("path", path))
	return os.Open(path)
}

func (bl *benchmarkList) readResults(ctx context.Context, reader io.ReadCloser) error {
	data := make(map[string]*pb.Benchmark)
	decoder := json.NewDecoder(reader)
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

	if len(cfg.URL) == 0 {
		ctxlog.G(ctx).Debug("benchmark list loading is disabled, skipping")
		return ls, nil
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
	URL string `yaml:"url"`
}
