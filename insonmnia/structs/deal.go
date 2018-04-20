package structs

import (
	"github.com/sonm-io/core/proto"
)

type DealID string

func (id DealID) String() string {
	return string(id)
}

type DealMeta struct {
	Deal     *sonm.Deal
	BidOrder *sonm.Order
	AskOrder *sonm.Order
	Tasks    []*TaskInfo
}

func NewDealMeta(d *sonm.Deal) *DealMeta {
	m := &DealMeta{
		Deal:  d,
		Tasks: make([]*TaskInfo, 0),
	}
	return m
}
