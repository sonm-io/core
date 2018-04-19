package structs

import (
	"github.com/sonm-io/core/proto"
)

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
