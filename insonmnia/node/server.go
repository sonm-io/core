package node

import (
	pb "github.com/sonm-io/core/proto"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	ListenAddr string
}

func Serve(c *Config) error {
	lis, err := net.Listen("tcp", c.ListenAddr)
	if err != nil {
		return err
	}

	deals := newDealsAPI()
	tasks := newTasksAPI()
	hub := newHubAPI()
	market := newMarketAPI()

	srv := grpc.NewServer()
	pb.RegisterDealManagementServer(srv, deals)
	pb.RegisterTaskManagementServer(srv, tasks)
	pb.RegisterHubManagementServer(srv, hub)
	pb.RegisterMarketServer(srv, market)

	return srv.Serve(lis)
}
