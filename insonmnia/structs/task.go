package structs

import (
	"errors"

	"github.com/sonm-io/core/proto"
)

var (
	errDealRequired  = errors.New("deal is required")
	errBidIdRequired = errors.New("bid id must be non-empty")
)

type StartTaskRequest struct {
	inner *sonm.HubStartTaskRequest
}

func NewStartTaskRequest(request *sonm.HubStartTaskRequest) (*StartTaskRequest, error) {
	deal := request.GetDeal()
	if deal == nil {
		return nil, errDealRequired
	}

	if deal.GetBidID() == "" {
		return nil, errBidIdRequired
	}

	return &StartTaskRequest{inner: request}, nil
}

func (r *StartTaskRequest) GetDeal() *sonm.Deal {
	return r.inner.GetDeal()
}

func (r *StartTaskRequest) GetOrderId() string {
	return r.inner.GetDeal().GetBidID()
}
