package node

import (
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type hubAPI struct {
	conf Config
	hub  pb.HubClient
	ctx  context.Context
}

func (h *hubAPI) Status(ctx context.Context, req *pb.Empty) (*pb.HubStatusReply, error) {
	log.G(h.ctx).Info("handling Status request")
	return h.hub.Status(ctx, req)
}

func (h *hubAPI) WorkersList(ctx context.Context, req *pb.Empty) (*pb.ListReply, error) {
	log.G(h.ctx).Info("handling WorkersList request")
	return h.hub.List(ctx, req)
}

func (h *hubAPI) WorkerStatus(ctx context.Context, req *pb.ID) (*pb.InfoReply, error) {
	log.G(h.ctx).Info("handling WorkersStatus request")
	return h.hub.Info(ctx, req)
}

func (h *hubAPI) GetRegisteredWorkers(ctx context.Context, req *pb.Empty) (*pb.GetRegisteredWorkersReply, error) {
	log.G(h.ctx).Info("handling GetRegisteredWorkers request")
	return h.hub.GetRegisteredWorkers(ctx, req)
}

func (h *hubAPI) RegisterWorker(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling RegisterWorkers request")
	return h.hub.RegisterWorker(ctx, req)
}

func (h *hubAPI) DeregisterWorker(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling DeregisterWorkers request")
	return h.hub.DeregisterWorker(ctx, req)
}

func (h *hubAPI) DeviceList(ctx context.Context, req *pb.Empty) (*pb.DevicesReply, error) {
	log.G(h.ctx).Info("handling DeviceList request")
	return h.hub.Devices(ctx, req)
}

func (h *hubAPI) GetDeviceProperties(ctx context.Context, req *pb.ID) (*pb.GetDevicePropertiesReply, error) {
	log.G(h.ctx).Info("handling GetDeviceProperties request")
	return h.hub.GetDeviceProperties(ctx, req)
}

func (h *hubAPI) SetDeviceProperties(ctx context.Context, req *pb.SetDevicePropertiesRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling SetDeviceProperties request")
	return h.hub.SetDeviceProperties(ctx, req)
}

func (h *hubAPI) GetAskPlan(context.Context, *pb.ID) (*pb.SlotsReply, error) {
	return &pb.SlotsReply{}, nil
}

func (h *hubAPI) GetAskPlans(ctx context.Context, req *pb.Empty) (*pb.SlotsReply, error) {
	log.G(h.ctx).Info("GetAskPlans")
	return h.hub.Slots(ctx, &pb.Empty{})
}

func (h *hubAPI) CreateAskPlan(ctx context.Context, req *pb.Slot) (*pb.Empty, error) {
	log.G(h.ctx).Info("CreateAskPlan")
	return h.hub.InsertSlot(ctx, req)
}

func (h *hubAPI) RemoveAskPlan(ctx context.Context, req *pb.Slot) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling RemoveAskPlan request")
	return h.hub.RemoveSlot(ctx, req)
}

func (h *hubAPI) TaskList(ctx context.Context, req *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling TaskList request")
	return h.hub.TaskList(ctx, &pb.Empty{})
}

func (h *hubAPI) TaskStatus(ctx context.Context, req *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(h.ctx).Info("handling TaskStatus request")
	return h.hub.TaskStatus(ctx, req)
}

func newHubAPI(ctx context.Context, conf Config) (pb.HubManagementServer, error) {
	cc, err := util.MakeGrpcClient(ctx, conf.HubEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	return &hubAPI{
		conf: conf,
		ctx:  ctx,
		hub:  pb.NewHubClient(cc),
	}, nil
}
