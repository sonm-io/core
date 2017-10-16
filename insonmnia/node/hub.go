package node

import (
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type hubAPI struct{}

func (h *hubAPI) Status(context.Context, *pb.Empty) (*pb.HubStatusReply, error) {
	return &pb.HubStatusReply{}, nil
}

func (h *hubAPI) WorkersList(context.Context, *pb.Empty) (*pb.ListReply, error) {
	return &pb.ListReply{}, nil
}

func (h *hubAPI) WorkersStatus(context.Context, *pb.HubInfoRequest) (*pb.InfoReply, error) {
	return &pb.InfoReply{}, nil
}

func (h *hubAPI) GetWorkerProperties(context.Context, *pb.ID) (*pb.GetMinerPropertiesReply, error) {
	return &pb.GetMinerPropertiesReply{}, nil
}

func (h *hubAPI) SetWorkerProperties(context.Context, *pb.SetMinerPropertiesRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (h *hubAPI) GetAskPlan(context.Context, *pb.ID) (*pb.GetSlotsReply, error) {
	return &pb.GetSlotsReply{}, nil
}

func (h *hubAPI) CreateAskPlan(context.Context, *pb.AddSlotRequest) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (h *hubAPI) RemoveAskPlan(context.Context, *pb.ID) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (h *hubAPI) TaskList(context.Context, *pb.Empty) (*pb.TaskListReply, error) {
	return &pb.TaskListReply{}, nil
}

func (h *hubAPI) TaskStatus(context.Context, *pb.ID) (*pb.TaskStatusReply, error) {
	return &pb.TaskStatusReply{}, nil
}

func newHubAPI() pb.HubManagementServer {
	return &hubAPI{}
}
