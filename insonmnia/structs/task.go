package structs

import (
	"errors"

	"github.com/sonm-io/core/proto"
)

var (
	errDealRequired   = errors.New("deal is required")
	errDealIdRequired = errors.New("deal id must be non-empty")
)

type StartTaskRequest struct {
	*sonm.StartTaskRequest
}

func NewStartTaskRequest(request *sonm.StartTaskRequest) (*StartTaskRequest, error) {
	deal := request.GetDeal()
	if deal == nil {
		return nil, errDealRequired
	}

	if deal.GetId().IsZero() {
		return nil, errDealIdRequired
	}

	return &StartTaskRequest{request}, nil
}

func (r *StartTaskRequest) GetDeal() *sonm.Deal {
	return r.Deal
}

func (r *StartTaskRequest) GetDealId() string {
	return r.GetDeal().GetId().Unwrap().String()
}
