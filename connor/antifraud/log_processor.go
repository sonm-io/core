package antifraud

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
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

type logParseFunc func(context.Context, *commonLogProcessor, io.Reader)

type commonLogProcessor struct {
	log          *zap.Logger
	cfg          *ProcessorConfig
	deal         *sonm.Deal
	taskID       string
	taskClient   sonm.TaskManagementClient
	hashrateEWMA metrics.EWMA
	startTime    time.Time

	hashrate    float64
	logParser   logParseFunc
	historyFile *os.File
}

func newLogProcessor(cfg *ProcessorConfig, log *zap.Logger, conn *grpc.ClientConn, deal *sonm.Deal, taskID string, lp logParseFunc) Processor {
	log = log.Named("task-logs").With(zap.String("task_id", taskID),
		zap.String("deal_id", deal.GetId().Unwrap().String()))

	return &commonLogProcessor{
		log:          log,
		cfg:          cfg,
		deal:         deal,
		taskID:       taskID,
		logParser:    lp,
		taskClient:   sonm.NewTaskManagementClient(conn),
		hashrateEWMA: metrics.NewEWMA1(),
		startTime:    time.Now(),
		hashrate:     float64(deal.Benchmarks.GPUEthHashrate()),
	}
}

func (m *commonLogProcessor) TaskQuality() (bool, float64) {
	accurate := m.startTime.Add(m.cfg.TaskWarmupDelay).Before(time.Now())
	desired := float64(m.deal.Benchmarks.GPUEthHashrate())
	actual := m.hashrateEWMA.Rate()
	return accurate, actual / desired
}

func (m *commonLogProcessor) TaskID() string {
	return m.taskID
}

func (m *commonLogProcessor) Run(ctx context.Context) error {
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
			m.hashrateEWMA.Update(int64(m.hashrate))
		case <-ewmaTick.C:
			m.hashrateEWMA.Tick()
		}
	}
}

func (m *commonLogProcessor) maybeOpenHistoryFile() error {
	if len(m.cfg.LogDir) == 0 {
		return fmt.Errorf("task logs saving is not configured")
	}

	fileName := fmt.Sprintf("%s_%s.log", m.deal.GetId().Unwrap().String(), m.taskID)
	fullPath := path.Join(m.cfg.LogDir, fileName)

	file, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	m.historyFile = file
	return nil
}

func (m *commonLogProcessor) maybeSaveLogLine(line string) {
	if m.historyFile != nil {
		if _, err := m.historyFile.WriteString(line + "\n"); err != nil {
			m.log.Warn("cannot write task log", zap.String("file", m.historyFile.Name()), zap.Error(err))
		}
	}
}

func (m *commonLogProcessor) fetchLogs(ctx context.Context) error {
	request := &sonm.TaskLogsRequest{
		Type:   sonm.TaskLogsRequest_STDOUT,
		Id:     m.taskID,
		Follow: true,
		DealID: m.deal.Id,
	}

	if err := m.maybeOpenHistoryFile(); err != nil {
		m.log.Warn("failed to open log file, task logs wouldn't be saved", zap.Error(err))
	} else {
		defer m.historyFile.Close()
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
			m.hashrate = 0.
			m.log.Warn("failed to fetch logs from the task")
			continue
		}

		logReader := sonm.NewLogReader(cli)
		reader, writer := io.Pipe()

		go m.logParser(ctx, m, reader)

		_, err = stdcopy.StdCopy(writer, writer, logReader)
		m.log.Warn("stop reading logs for task", zap.Error(err))
		writer.Close()
	}
}

func newClaymoreLogProcessor(cfg *ProcessorConfig, log *zap.Logger, nodeConnection *grpc.ClientConn, deal *sonm.Deal, taskID string) Processor {
	return newLogProcessor(cfg, log, nodeConnection, deal, taskID, claymoreLogParser)
}

func claymoreLogParser(ctx context.Context, m *commonLogProcessor, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if ctx.Err() != nil {
			m.log.Debug("stop reading logs: context cancelled")
			return
		}

		line := scanner.Text()
		m.maybeSaveLogLine(line)

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

			m.hashrate = hashrate * 1e6
		}
	}

	m.log.Debug("finished reading logs", zap.Error(scanner.Err()))
}
