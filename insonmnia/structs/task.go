package structs

import (
	"errors"

	"github.com/sonm-io/core/proto"
)

var (
	errDealRequired  = errors.New("deal is required")
	errBidIdRequired = errors.New("buyer id must be non-empty")
)

type StartTaskRequest struct {
	*sonm.HubStartTaskRequest
}

func NewStartTaskRequest(request *sonm.HubStartTaskRequest) (*StartTaskRequest, error) {
	deal := request.GetDeal()
	if deal == nil {
		return nil, errDealRequired
	}

	if deal.GetBuyerID() == "" {
		return nil, errBidIdRequired
	}

	return &StartTaskRequest{request}, nil
}

func (r *StartTaskRequest) GetDeal() *sonm.Deal {
	return r.Deal
}

func (r *StartTaskRequest) GetOrderId() string {
	return r.GetDeal().GetId()
}
