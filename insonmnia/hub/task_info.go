package hub

import (
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
)

type TaskInfo struct {
	structs.StartTaskRequest
	pb.MinerStartReply
	ID      string
	MinerId string
}

type DealMeta struct {
	BidID string
	Order structs.Order
	Tasks []*TaskInfo
}
