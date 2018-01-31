package hub

import (
	"time"

	"github.com/sonm-io/core/insonmnia/resource"
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

func (t TaskInfo) ContainerID() string {
	return t.MinerStartReply.Container
}

type DealMeta struct {
	ID      DealID
	BidID   string
	MinerID string
	Order   structs.Order
	Usage   resource.Resources
	Tasks   []*TaskInfo
	EndTime time.Time
}
