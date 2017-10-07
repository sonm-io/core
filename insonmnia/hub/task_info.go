package hub

import (
	pb "github.com/sonm-io/core/proto"
)

type TaskInfo struct {
	pb.HubStartTaskRequest
	pb.MinerStartReply
	ID      string
	MinerId string
}
