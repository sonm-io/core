package optimus

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type namedErrorGroup struct {
	errs map[string]error
}

func newNamedErrorGroup() *namedErrorGroup {
	return &namedErrorGroup{
		errs: map[string]error{},
	}
}

func (m *namedErrorGroup) Set(id string, err error) {
	m.errs[id] = err
}

// SetUnique associates the given error with provided "ids" only and if only
// there wasn't an error associated with the "id" previously.
func (m *namedErrorGroup) SetUnique(ids []string, err error) {
	for _, id := range ids {
		if _, ok := m.errs[id]; !ok {
			m.errs[id] = err
		}
	}
}

func (m *namedErrorGroup) Error() string {
	errs := map[string]string{}
	for id, err := range m.errs {
		errs[id] = err.Error()
	}

	data, err := json.Marshal(errs)
	if err != nil {
		panic(fmt.Sprintf("failed to dump `namedErrorGroup` into JSON: %v", err))
	}

	return string(data)
}

func (m *namedErrorGroup) ErrorOrNil() error {
	if len(m.errs) == 0 {
		return nil
	}

	return m
}

type WorkerManagementClientAPI interface {
	Devices(ctx context.Context, request *sonm.Empty, opts ...grpc.CallOption) (*sonm.DevicesReply, error)
	AskPlans(ctx context.Context, request *sonm.Empty, opts ...grpc.CallOption) (*sonm.AskPlansReply, error)
	CreateAskPlan(ctx context.Context, request *sonm.AskPlan, opts ...grpc.CallOption) (*sonm.ID, error)
	RemoveAskPlan(ctx context.Context, request *sonm.ID, opts ...grpc.CallOption) (*sonm.Empty, error)
	NextMaintenance(ctx context.Context, request *sonm.Empty, opts ...grpc.CallOption) (*sonm.Timestamp, error)
}

type workerManagementClientAPI struct {
	WorkerManagementClientAPI
}

func (m *workerManagementClientAPI) CreateAskPlan(ctx context.Context, request *sonm.AskPlan, opts ...grpc.CallOption) (*sonm.ID, error) {
	// We need to clean this, because otherwise worker rejects such request.
	request.OrderID = nil

	return m.WorkerManagementClientAPI.CreateAskPlan(ctx, request, opts...)
}

// WorkerManagementClientExt extends default "WorkerManagementClient" with an
// ability to remove multiple ask-plans.
type WorkerManagementClientExt interface {
	WorkerManagementClientAPI
	RemoveAskPlans(ctx context.Context, ids []string) error
}

type workerManagementClientExt struct {
	WorkerManagementClientAPI
}

func (m *workerManagementClientExt) RemoveAskPlans(ctx context.Context, ids []string) error {
	errs := newNamedErrorGroup()

	// ID set for fast detection
	idSet := map[string]struct{}{}
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Concurrently remove all required ask-plans. The wait-group here always
	// returns nil.
	wg, ctx := errgroup.WithContext(ctx)
	for _, id := range ids {
		id := id
		wg.Go(func() error {
			if _, err := m.RemoveAskPlan(ctx, &sonm.ID{Id: id}); err != nil {
				errs.Set(id, err)
			}

			return nil
		})
	}
	wg.Wait()

	// Wait for ask plans be REALLY removed.
	timer := util.NewImmediateTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			errs.SetUnique(ids, ctx.Err())
			return errs
		case <-timer.C:
			plans, err := m.AskPlans(ctx, &sonm.Empty{})
			if err != nil {
				errs.SetUnique(ids, err)
				return errs
			}

			// Detecting set intersection.
			intersects := false
			for id := range plans.AskPlans {
				// Continue to wait if there are ask plans left.
				if _, ok := idSet[id]; ok {
					intersects = true
					break
				}
			}

			if !intersects {
				return errs.ErrorOrNil()
			}
		}
	}
}

// ReadOnlyWorker is a worker management client wrapper that allows only
// immutable operations. It returns some default response for operations that
// mutates something.
type ReadOnlyWorker struct {
	WorkerManagementClientAPI

	mu           sync.Mutex
	removedPlans map[string]struct{}
}

func NewReadOnlyWorker(worker WorkerManagementClientAPI) *ReadOnlyWorker {
	return &ReadOnlyWorker{
		WorkerManagementClientAPI: worker,

		removedPlans: map[string]struct{}{},
	}
}

func (m *ReadOnlyWorker) CreateAskPlan(ctx context.Context, in *sonm.AskPlan, opts ...grpc.CallOption) (*sonm.ID, error) {
	return &sonm.ID{Id: "00000000-0000-0000-0000-000000000000"}, nil
}

func (m *ReadOnlyWorker) AskPlans(ctx context.Context, in *sonm.Empty, opts ...grpc.CallOption) (*sonm.AskPlansReply, error) {
	plans, err := m.WorkerManagementClientAPI.AskPlans(ctx, in, opts...)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	response := &sonm.AskPlansReply{
		AskPlans: map[string]*sonm.AskPlan{},
	}

	for k, v := range plans.AskPlans {
		if _, ok := m.removedPlans[k]; !ok {
			response.AskPlans[k] = v
		}
	}

	return response, nil
}

func (m *ReadOnlyWorker) RemoveAskPlan(ctx context.Context, in *sonm.ID, opts ...grpc.CallOption) (*sonm.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.removedPlans[in.Id] = struct{}{}

	return &sonm.Empty{}, nil
}

func (m *ReadOnlyWorker) PurgeAskPlans(ctx context.Context, in *sonm.Empty, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

func (m *ReadOnlyWorker) ScheduleMaintenance(ctx context.Context, in *sonm.Timestamp, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

func (m *ReadOnlyWorker) RemoveBenchmark(ctx context.Context, in *sonm.NumericID, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

func (m *ReadOnlyWorker) PurgeBenchmarks(ctx context.Context, in *sonm.Empty, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

type mockWorker struct {
	PredefinedDevices *sonm.DevicesReply
	Result            []*sonm.AskPlan
}

func newMockWorker(devices *sonm.DevicesReply) *mockWorker {
	return &mockWorker{
		PredefinedDevices: devices,
	}
}

func (m *mockWorker) Devices(ctx context.Context, request *sonm.Empty, opts ...grpc.CallOption) (*sonm.DevicesReply, error) {
	return m.PredefinedDevices, nil
}

func (m *mockWorker) AskPlans(ctx context.Context, request *sonm.Empty, opts ...grpc.CallOption) (*sonm.AskPlansReply, error) {
	return &sonm.AskPlansReply{}, nil
}

func (m *mockWorker) CreateAskPlan(ctx context.Context, request *sonm.AskPlan, opts ...grpc.CallOption) (*sonm.ID, error) {
	m.Result = append(m.Result, request)
	return &sonm.ID{Id: "00000000-0000-0000-0000-000000000000"}, nil
}

func (m *mockWorker) RemoveAskPlan(ctx context.Context, request *sonm.ID, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return &sonm.Empty{}, nil
}

func (m *mockWorker) NextMaintenance(ctx context.Context, request *sonm.Empty, opts ...grpc.CallOption) (*sonm.Timestamp, error) {
	return &sonm.Timestamp{Seconds: math.MaxInt32}, nil
}
