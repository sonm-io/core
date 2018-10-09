package resource

import (
	"context"
	"fmt"
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v1"
)

type Scheduler struct {
	OS   *hardware.Hardware
	mu   sync.Mutex
	pool *pool
	// taskToAskPlan maps task ID to ask plan ID
	taskToAskPlan map[string]string
	// askPlanPools maps ask plan' ID to allocated resource pool
	askPlanPools map[string]*taskPool
	log          *zap.SugaredLogger
}

func NewScheduler(ctx context.Context, hardware *hardware.Hardware) *Scheduler {
	resources := hardware.AskPlanResources()
	log := ctxlog.S(ctx).With("source", "resource_scheduler")
	readableResources, _ := yaml.Marshal(resources)
	readableHardware, _ := yaml.Marshal(hardware)
	log.Debugf("constructing scheduler with hardware:\n%s\ninitial resources:\n%s", string(readableHardware), string(readableResources))
	return &Scheduler{
		OS:            hardware,
		pool:          newPool(log, resources),
		taskToAskPlan: map[string]string{},
		askPlanPools:  map[string]*taskPool{},
		log:           log,
	}
}

func (m *Scheduler) DebugDump() *sonm.SchedulerData {
	m.mu.Lock()
	defer m.mu.Unlock()
	panic("implement me")

	//reply := &sonm.SchedulerData{
	//	TaskToAskPlan: deepcopy.Copy(m.taskToAskPlan).(map[string]string),
	//	MainPool: &sonm.ResourcePool{
	//		All:  m.pool.all,
	//		Used: map[string]*sonm.AskPlanResources{},
	//	},
	//	AskPlanPools: map[string]*sonm.ResourcePool{},
	//}
	//
	//for id, res := range m.pool.used {
	//	reply.MainPool.Used[id] = res
	//}
	//
	//for askID, pool := range m.askPlanPools {
	//	resultPool := &sonm.ResourcePool{
	//		All:  pool.all,
	//		Used: map[string]*sonm.AskPlanResources{},
	//	}
	//	reply.AskPlanPools[askID] = resultPool
	//	for id, res := range pool.used {
	//		resultPool.Used[id] = res
	//	}
	//}
	//
	//return reply
}

//TODO: rework needed — looks like it should not be here
func (m *Scheduler) AskPlanIDByTaskID(taskID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	askID, ok := m.taskToAskPlan[taskID]
	if !ok {
		return "", fmt.Errorf("ask plan for task %s is not found", taskID)
	}
	return askID, nil
}

func (m *Scheduler) GetCommitedFree() (*sonm.AskPlanResources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool.getCommitedFree()
}

// Consume tries to consume the specified resource usage from the pool.
//
// Does nothing on error.
func (m *Scheduler) Consume(askPlan *sonm.AskPlan) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.askPlanPools[askPlan.ID] = newTaskPool(askPlan.Resources)

	if err := m.pool.consume(askPlan); err != nil {
		return fmt.Errorf("failed to consume resources for ask plan %s: %s", askPlan.ID, err)
	}
	m.log.Debugf("consumed ask-plan %s by scheduler", askPlan.ID)
	return nil
}

func (m *Scheduler) MakeRoomAndCommit(askPlan *sonm.AskPlan) (ejectedAskPlans []string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool.makeRoomAndCommit(askPlan)
}

func (m *Scheduler) ConsumeTask(askPlanID string, taskID string, resources *sonm.AskPlanResources) error {
	copy := &sonm.AskPlanResources{
		GPU:     deepcopy.Copy(resources.GetGPU()).(*sonm.AskPlanGPU),
		Storage: deepcopy.Copy(resources.GetStorage()).(*sonm.AskPlanStorage),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.taskToAskPlan[taskID] = askPlanID
	pool, ok := m.askPlanPools[askPlanID]
	if !ok {
		return fmt.Errorf("could not consume resources for task - ask Plan with id %s not found", askPlanID)
	}
	m.log.Debugf("consumed task %s by scheduler", taskID)

	return pool.consume(taskID, copy)
}

func (m *Scheduler) Release(askPlanID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.askPlanPools, askPlanID)
	if err := m.pool.release(askPlanID); err != nil {
		return fmt.Errorf("failed to release ask plan %s from scheduler: %s", askPlanID, err)
	}
	m.log.Debugf("released ask plan %s from scheduler", askPlanID)
	return nil
}

func (m *Scheduler) ReleaseTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	askPlanID, ok := m.taskToAskPlan[taskID]
	if !ok {
		return fmt.Errorf("failed to release task %s from scheduler: could not find corresponding ask plan", taskID)
	}
	pool, ok := m.askPlanPools[askPlanID]
	if !ok {
		return fmt.Errorf("failed to release task %s: ask Plan with id %s not found", taskID, askPlanID)
	}

	err := pool.release(taskID)
	if err != nil {
		return err
	}
	m.log.Debugf("released task %s", taskID)
	return nil
}

func (m *Scheduler) ResourceByTask(taskID string) (*sonm.AskPlanResources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	askID, ok := m.taskToAskPlan[taskID]
	if !ok {
		return nil, fmt.Errorf("failed to get ask plan id for task %s: no such ask plan", taskID)
	}

	pool, ok := m.askPlanPools[askID]
	if !ok {
		return nil, fmt.Errorf("failed to get ask plan pool by id %s: no such pool", askID)
	}

	res, ok := pool.used[taskID]
	if !ok {
		return nil, fmt.Errorf("failed to get resources for task %s: no such task", taskID)
	}

	return res, nil
}

func (m *Scheduler) OnDealFinish(taskID string) error {
	// We don't care about error here if the task was already released, but we need to be sure, if it was not.
	m.ReleaseTask(taskID)

	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.taskToAskPlan[taskID]

	if !ok {
		return fmt.Errorf("failed finish deal for task %s from scheduler: could not find corresponding ask plan", taskID)
	}

	delete(m.taskToAskPlan, taskID)

	return nil
}
