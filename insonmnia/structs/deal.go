package structs

import (
	"github.com/sonm-io/core/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DealRequest struct {
	*sonm.DealRequest
}

func NewDealRequest(deal *sonm.DealRequest) (*DealRequest, error) {
	if deal.GetBidId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "bid_id is required")
	}

	if deal.GetAskId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ask_id is required")
	}

	if deal.GetSpecHash() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "specification hash is required")
	}

	return &DealRequest{deal}, nil
}

type DealID string

func (id DealID) String() string {
	return string(id)
}

type DealMeta struct {
	Deal     *sonm.MarketDeal
	BidOrder *sonm.MarketOrder
	AskOrder *sonm.MarketOrder
	Tasks    []*TaskInfo
}

func NewDealMeta(d *sonm.MarketDeal) *DealMeta {
	m := &DealMeta{
		Deal:  d,
		Tasks: make([]*TaskInfo, 0),
	}
	return m
}
