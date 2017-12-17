package hub

import (
	"time"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
)

type TaskInfo struct {
	structs.StartTaskRequest
	pb.MinerStartReply
	ID      string
	DealId  DealID
	MinerId string
	EndTime *time.Time
}

type DealMeta struct {
	BidID string
	Order structs.Order
	Tasks []*TaskInfo
}
