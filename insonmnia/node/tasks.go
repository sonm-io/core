package node

import (
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type tasksAPI struct{}

func (t *tasksAPI) List(context.Context, *pb.TaskListRequest) (*pb.TaskListReply, error) {
	return &pb.TaskListReply{}, nil
}
func (t *tasksAPI) Start(context.Context, *pb.HubStartTaskRequest) (*pb.TaskInfo, error) {
	return &pb.TaskInfo{}, nil
}
func (t *tasksAPI) Status(context.Context, *pb.ID) (*pb.TaskInfo, error) {
	return &pb.TaskInfo{}, nil
}
func (t *tasksAPI) Logs(*pb.ID, pb.TaskManagement_LogsServer) error {
	return nil
}

func (t *tasksAPI) Stop(context.Context, *pb.ID) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func newTasksAPI() pb.TaskManagementServer {
	return &tasksAPI{}
}
