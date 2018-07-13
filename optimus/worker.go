package optimus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sonm-io/core/proto"
	"golang.org/x/sync/errgroup"
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

// WorkerManagementClientExt extends default "WorkerManagementClient" with an
// ability to remove multiple ask-plans.
type WorkerManagementClientExt interface {
	sonm.WorkerManagementClient
	RemoveAskPlans(ctx context.Context, ids []string) error
}

type workerManagementClientExt struct {
	sonm.WorkerManagementClient
}

func (m *workerManagementClientExt) RemoveAskPlans(ctx context.Context, ids []string) error {
	errs := newNamedErrorGroup()

	// ID set for fast detection
	idSet := map[string]struct{}{}
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	// Concurrently remove all required ask-plans. The wait-group here always
	// returns nil.
	wg := errgroup.Group{}
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
	timer := time.NewTicker(1 * time.Second)
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
