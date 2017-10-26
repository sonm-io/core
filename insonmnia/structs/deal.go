package structs

import (
	"github.com/sonm-io/core/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrOrderIsRequired = status.Errorf(codes.InvalidArgument, "order is required")
	ErrSlotIsRequired  = status.Errorf(codes.InvalidArgument, "slot is required")
)

type DealRequest struct {
	*sonm.DealRequest
}

func NewDealRequest(deal *sonm.DealRequest) (*DealRequest, error) {
	order := deal.GetOrder()
	if order == nil {
		return nil, ErrOrderIsRequired
	}
	slot := order.GetSlot()
	if slot == nil {
		return nil, ErrSlotIsRequired
	}

	return &DealRequest{deal}, nil
}
