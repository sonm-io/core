package antifraud

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rcrowley/go-metrics"
	"github.com/sonm-io/core/connor/price"
	"github.com/sonm-io/core/connor/types"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"gopkg.in/oleiade/lane.v1"
)

type dwarfPoolProcessor struct {
	cfg      *ProcessorConfig
	log      *zap.Logger
	wallet   common.Address
	taskID   string
	workerID string
	deal     *types.Deal

	startTime       time.Time
	currentHashrate float64
	hashrateEWMA    metrics.EWMA
	hashrateQueue   *lane.Queue
}

func newDwarfPoolProcessor(cfg *ProcessorConfig, log *zap.Logger, deal *types.Deal, taskID string) *dwarfPoolProcessor {
	workerID := fmt.Sprintf("c%s", deal.GetId().Unwrap().String())
	l := log.Named("dwarfpool").With(
		zap.String("deal_id", deal.GetId().Unwrap().String()),
		zap.String("task_id", taskID),
		zap.String("worker_id", workerID))

	return &dwarfPoolProcessor{
		log:             l,
		cfg:             cfg,
		taskID:          taskID,
		workerID:        workerID,
		wallet:          deal.GetConsumerID().Unwrap(),
		startTime:       time.Now(),
		deal:            deal,
		hashrateEWMA:    metrics.NewEWMA(1 - math.Exp(-5.0/cfg.DecayTime)),
		currentHashrate: float64(deal.BenchmarkValue()),
		hashrateQueue:   &lane.Queue{Deque: lane.NewCappedDeque(60)},
	}
}

func (w *dwarfPoolProcessor) Run(ctx context.Context) error {
	timer := time.NewTimer(w.cfg.TaskWarmupDelay)
	w.log.Info("starting task's warm-up")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		w.log.Debug("warm-up complete, starting watcher", zap.String("wallet", w.wallet.Hex()))
	}
	timer.Stop()

	// This should not be configured, as ticker in ewma is bound to 5 seconds
	ewmaTick := time.NewTicker(5 * time.Second)
	ewmaUpdate := time.NewTicker(1 * time.Second)
	track := util.NewImmediateTicker(w.cfg.TrackInterval)

	defer ewmaUpdate.Stop()
	defer ewmaTick.Stop()
	defer track.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ewmaUpdate.C:
			w.hashrateEWMA.Update(int64(w.currentHashrate))
		case <-ewmaTick.C:
			w.hashrateEWMA.Tick()
		case <-track.C:
			if err := w.watch(); err != nil {
				w.log.Warn("failed to load dwarfPool's data", zap.Error(err))
			}
		}
	}
}

func (w *dwarfPoolProcessor) TaskID() string {
	return w.taskID
}

func (w *dwarfPoolProcessor) TaskQuality() (bool, float64) {
	// should not be configured - ewma is bound to 1 hour rate
	accurate := w.startTime.Add(time.Hour).Before(time.Now())
	desired := float64(w.deal.BenchmarkValue())
	actual := w.hashrateEWMA.Rate()
	rate := actual / desired

	if !w.nonZeroHashrate() {
		rate = 0
	}

	return accurate, rate
}

func (w *dwarfPoolProcessor) watch() error {
	url := fmt.Sprintf("http://dwarfpool.com/eth/api?wallet=%s", strings.ToLower(w.wallet.Hex()))
	data, err := price.FetchURLWithRetry(url)
	if err != nil {
		return fmt.Errorf("failed to fetch dwarfpool data: %v", err)
	}

	resp := &dwarfPoolResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return fmt.Errorf("failed to parse dwarfpool response: %v", err)
	}

	worker, ok := resp.Workers[w.workerID]
	if !ok {
		return fmt.Errorf("cannot find worker %s in reponse data", w.workerID)
	}

	w.log.Info("task hashrate",
		zap.Float64("reported", worker.Hashrate),
		zap.Float64("calculated", worker.HashrateCalculated))

	var rate float64
	if worker.HashrateCalculated > 0 {
		rate = worker.HashrateCalculated
	} else {
		rate = worker.Hashrate
	}

	w.currentHashrate = rate * 1e6
	w.updateHashRateQueue(w.currentHashrate)
	return nil
}

func (w *dwarfPoolProcessor) updateHashRateQueue(v float64) {
	if !w.hashrateQueue.Full() {
		w.hashrateQueue.Append(v)
	} else {
		w.hashrateQueue.Shift()
		w.hashrateQueue.Append(v)
	}
}

func (w *dwarfPoolProcessor) nonZeroHashrate() bool {
	if w.hashrateQueue.Size() < 5 {
		return true
	}

	for i := 0; i < 5; i++ {
		v := w.hashrateQueue.Pop()
		w.hashrateQueue.Append(v)
		if v.(float64) > 0 {
			return true
		}
	}

	return false
}

type dwarfPoolWorker struct {
	Alive              bool    `json:"alive"`
	Hashrate           float64 `json:"hashrate"`
	HashrateCalculated float64 `json:"hashrate_calculated"`
	SecondSinceSubmit  int64   `json:"second_since_submit"`
}

type dwarfPoolResponse struct {
	Workers map[string]*dwarfPoolWorker `json:"workers"`
}
