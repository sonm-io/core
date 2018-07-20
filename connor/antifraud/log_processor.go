package antifraud

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rcrowley/go-metrics"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewLogProcessor(cfg *LogProcessorConfig, log *zap.Logger, nodeConnection *grpc.ClientConn, deal *sonm.Deal, taskID string) LogProcessor {
	taskLogger := log.Named("task-logs").With(zap.String("task_id", taskID), zap.String("deal_id", deal.GetId().Unwrap().String()))
	return &EthClaymoreLogProcessor{
		cfg:              cfg,
		log:              taskLogger,
		deal:             deal,
		taskID:           taskID,
		taskClient:       sonm.NewTaskManagementClient(nodeConnection),
		hashrateEWMA:     metrics.NewEWMA1(),
		lastHashrateTime: time.Now(),
		startTime:        time.Now(),
		currentHashrate:  float64(deal.Benchmarks.GPUEthHashrate()),
	}
}

type NilLogProcessor struct{}

func (m *NilLogProcessor) TaskQuality() float64 {
	return 1.
}

func (m *NilLogProcessor) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

type EthClaymoreLogProcessor struct {
	log          *zap.Logger
	cfg          *LogProcessorConfig
	deal         *sonm.Deal
	taskID       string
	taskClient   sonm.TaskManagementClient
	hashrateEWMA metrics.EWMA

	startTime        time.Time
	lastHashrateTime time.Time
	currentHashrate  float64
}

func (m *EthClaymoreLogProcessor) TaskQuality() (bool, float64) {
	accurate := m.startTime.Add(m.cfg.TaskWarmupDelay).Before(time.Now())
	desired := float64(m.deal.Benchmarks.GPUEthHashrate())
	actual := m.hashrateEWMA.Rate()
	return accurate, actual / desired
}

func (m *EthClaymoreLogProcessor) TaskID() string {
	return m.taskID
}

func (m *EthClaymoreLogProcessor) Run(ctx context.Context) error {
	go m.fetchLogs(ctx)
	m.log.Info("starting task's warm-up")
	timer := time.NewTimer(m.cfg.TaskWarmupDelay)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	timer.Stop()
	m.log.Info("task is warmed-up")

	// This should not be configured, as ticker in ewma is bound to 5 seconds
	ewmaTick := time.NewTicker(5 * time.Second)
	ewmaUpdate := time.NewTicker(1 * time.Second)

	defer ewmaTick.Stop()
	defer ewmaUpdate.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ewmaUpdate.C:
			m.hashrateEWMA.Update(int64(m.currentHashrate))
		case <-ewmaTick.C:
			m.hashrateEWMA.Tick()
		}
	}
}

//TODO: get rid of the copy paste!!!!! see cmd/cli/commands/tasks.go
type logReader struct {
	cli      sonm.TaskManagement_LogsClient
	buf      bytes.Buffer
	finished bool
}

func (m *logReader) Read(p []byte) (n int, err error) {
	if len(p) > m.buf.Len() && !m.finished {
		chunk, err := m.cli.Recv()
		if err == io.EOF {
			m.finished = true
		} else if err != nil {
			return 0, err
		}
		if chunk != nil && chunk.Data != nil {
			m.buf.Write(chunk.Data)
		}
	}
	return m.buf.Read(p)
}

func (m *EthClaymoreLogProcessor) fetchLogs(ctx context.Context) error {
	request := &sonm.TaskLogsRequest{
		Type:   sonm.TaskLogsRequest_STDOUT,
		Id:     m.taskID,
		Follow: true,
		DealID: m.deal.Id,
	}

	retryTicker := util.NewImmediateTicker(m.cfg.TrackInterval)
	defer retryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retryTicker.C:
		}

		m.log.Debug("requesting logs")
		cli, err := m.taskClient.Logs(ctx, request)
		if err != nil {
			m.currentHashrate = 0.
			m.log.Warn("failed to fetch logs from the task")
			continue
		}

		logReader := &logReader{cli: cli}
		reader, writer := io.Pipe()

		go m.parseLogs(ctx, reader)

		_, err = stdcopy.StdCopy(writer, writer, logReader)
		m.log.Warn("stop reading logs for task", zap.Error(err))
		writer.Close()
	}
}

func (m *EthClaymoreLogProcessor) parseLogs(ctx context.Context, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Total Speed: ") {
			fields := strings.Fields(line)
			if len(fields) != 13 {
				m.log.Warn("invalid claymore log line", zap.String("line", line), zap.Int("fields_count", len(fields)))
				return
			}

			hashrateStr := fields[4]
			hashrate, err := strconv.ParseFloat(hashrateStr, 64)
			if err != nil {
				m.log.Warn("failed to parse hashrate",
					zap.String("line", line),
					zap.String("field", hashrateStr),
					zap.Error(err))
				return
			}

			m.currentHashrate = hashrate * 1e6
			m.lastHashrateTime = time.Now()
		}
	}

	m.log.Warn("finished reading logs", zap.Error(scanner.Err()))
}
