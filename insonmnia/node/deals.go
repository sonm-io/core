package node

import (
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type dealsAPI struct{}

func (d *dealsAPI) List(context.Context, *pb.DealListRequest) (*pb.DealListReply, error) {
	return &pb.DealListReply{
		Deal: []*pb.Deal{},
	}, nil
}

func (d *dealsAPI) Status(context.Context, *pb.ID) (*pb.Deal, error) {
	return &pb.Deal{}, nil
}

func (d *dealsAPI) Finish(context.Context, *pb.ID) (*pb.Empty, error) {
	return nil, nil
}

func newDealsAPI() pb.DealManagementServer {
	return &dealsAPI{}
}
