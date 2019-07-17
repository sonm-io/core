package action

import (
	"context"
)

// Action is an abstraction of some action that can be rolled back.
type Action interface {
	// Execute executes this action, returning error if any.
	Execute(ctx context.Context) error
	// Rollback rollbacks this action, returning error if any.
	Rollback() error
}
