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

func (h *hubAPI) GetRegistredWorkers(ctx context.Context, req *pb.Empty) (*pb.GetRegistredWorkersReply, error) {
	log.G(h.ctx).Info("handling GetRegistredWorkers request")
	return h.hub.GetRegistredWorkers(ctx, req)
}

func (h *hubAPI) RegisterWorker(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling RegisterWorkers request")
	return h.hub.RegisterWorker(ctx, req)
}

func (h *hubAPI) UnregisterWorker(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling UnregisterWorkers request")
	return h.hub.UnregisterWorker(ctx, req)
}

func (h *hubAPI) GetWorkerProperties(ctx context.Context, req *pb.ID) (*pb.GetMinerPropertiesReply, error) {
	log.G(h.ctx).Info("handling GetWorkerProperties request")
	return h.hub.GetMinerProperties(ctx, req)
}

func (h *hubAPI) SetWorkerProperties(ctx context.Context, req *pb.SetMinerPropertiesRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling SetWorkerProperties request")
	return h.hub.SetMinerProperties(ctx, req)
}

func (h *hubAPI) GetAskPlans(ctx context.Context, req *pb.Empty) (*pb.GetAllSlotsReply, error) {
	log.G(h.ctx).Info("GetAskPlan")
	reply, err := h.hub.GetAllSlots(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (h *hubAPI) CreateAskPlan(ctx context.Context, req *pb.AddSlotRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("CreateAskPlan")
	return h.hub.AddSlot(ctx, req)
}

func (h *hubAPI) RemoveAskPlan(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("RemoveAskPlan")
	request := &pb.RemoveSlotRequest{ID: req.GetId()}
	return h.hub.RemoveSlot(ctx, request)
}

func (h *hubAPI) TaskList(ctx context.Context, req *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling TaskList request")
	return h.hub.TaskList(ctx, &pb.Empty{})
}

func (h *hubAPI) TaskStatus(ctx context.Context, req *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(h.ctx).Info("handling TaskStatus request")
	return h.hub.TaskStatus(ctx, req)
}

func (h *hubAPI) getWorkersIDs(ctx context.Context) ([]string, error) {
	workers, err := h.hub.List(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for id := range workers.GetInfo() {
		ids = append(ids, id)
	}

	return ids, nil
}

func newHubAPI(ctx context.Context, conf Config) (pb.HubManagementServer, error) {
	cc, err := util.MakeGrpcClient(conf.HubEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	return &hubAPI{
		conf: conf,
		ctx:  ctx,
		hub:  pb.NewHubClient(cc),
	}, nil
}
