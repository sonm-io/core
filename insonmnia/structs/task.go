package structs

import (
	"errors"
	"time"

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

	if deal.GetId() == "" {
		return nil, errDealIdRequired
	}

	return &StartTaskRequest{request}, nil
}

func (r *StartTaskRequest) GetDeal() *sonm.Deal {
	return r.Deal
}

func (r *StartTaskRequest) GetDealId() *sonm.BigInt {
	v, _ := sonm.NewBigIntFromString(r.GetDeal().GetId())
	return v
}

type TaskInfo struct {
	StartTaskRequest
	sonm.MinerStartReply
	ID      string
	DealId  DealID
	MinerId string
	EndTime *time.Time
}

func (t TaskInfo) ContainerID() string {
	return t.MinerStartReply.Container
}
