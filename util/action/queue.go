package action

import (
	"context"

	"github.com/sonm-io/core/util/multierror"
)

// ActionQueue represents a queue of executable actions.
// Any action that fails triggers cascade previous actions rollback.
type ActionQueue struct {
	actions []Action
}

func NewActionQueue(actions ...Action) *ActionQueue {
	return &ActionQueue{
		actions: actions,
	}
}

// Execute executes the action queue.
//
// If any of actions fails a cascade previous actions rollback occurs resulting
// in a tuple of this's action error and rollback ones if any.
func (m *ActionQueue) Execute(ctx context.Context) (error, error) {
	executedActions := make([]Action, 0)
	for _, action := range m.actions {
		if err := action.Execute(ctx); err != nil {
			return err, Rollback(executedActions)
		}

		executedActions = append(executedActions, action)
	}

	return nil, nil
}

func Rollback(actions []Action) error {
	queue := &deque{actions: actions}

	errs := multierror.NewMultiError()
	for {
		if action, ok := queue.pop(); ok {
			errs = multierror.Append(errs, action.Rollback())
		} else {
			break
		}
	}

	return errs.ErrorOrNil()
}

type deque struct {
	actions []Action
}

func (m *deque) pop() (Action, bool) {
	length := len(m.actions)
	if length == 0 {
		return nil, false
	}

	action := m.actions[length-1]
	m.actions = m.actions[:length-1]
	return action, true
}
