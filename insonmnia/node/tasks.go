package node

import (
	"crypto/ecdsa"
	"io"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

type tasksAPI struct {
	key     *ecdsa.PrivateKey
	bc      *blockchain.API
	locator pb.LocatorClient
	ctx     context.Context
	creds   credentials.TransportCredentials
}

func (t *tasksAPI) List(ctx context.Context, req *pb.TaskListRequest) (*pb.TaskListReply, error) {
	log.G(t.ctx).Info("handling List request", zap.Any("request", req))

	// has hubID, can perform direct request
	if req.GetHubID() != "" {
		log.G(t.ctx).Info("has HubID, performing direct request")
		hubClient, err := t.getHubClientByEthAddr(ctx, req.GetHubID())
		if err != nil {
			return nil, err
		}

		return hubClient.TaskList(ctx, &pb.Empty{})
	}

	myAddr := util.PubKeyToAddr(t.key.PublicKey)
	dealIDs, err := t.bc.GetDeals(myAddr)
	if err != nil {
		return nil, err
	}

	log.G(t.ctx).Info("found some deals", zap.Int("count", len(dealIDs)))

	var activeDeals []*pb.Deal
	for _, id := range dealIDs {
		dealInfo, err := t.bc.GetDealInfo(id)
		if err != nil {
			return nil, err
		}

		// NOTE: status id should be changed
		if dealInfo.Status == pb.DealStatus_ACCEPTED {
			activeDeals = append(activeDeals, dealInfo)
		}
	}

	log.G(t.ctx).Info("found some active deals", zap.Int("count", len(activeDeals)))

	tasks := make(map[string]*pb.TaskListReply_TaskInfo)
	for _, deal := range activeDeals {
		hub, err := t.getHubClientByEthAddr(ctx, deal.GetSupplierID())
		if err != nil {
			return nil, err
		}

		taskList, err := hub.TaskList(ctx, &pb.Empty{})
		if err != nil {
			return nil, err
		}

		for k, v := range taskList.GetInfo() {
			tasks[k] = v
		}
	}

	return &pb.TaskListReply{Info: tasks}, nil
}

func (t *tasksAPI) Start(ctx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	log.G(t.ctx).Info("handling Start request", zap.Any("request", req))

	hub, err := t.getHubClientForDeal(ctx, req.Deal.GetId())
	if err != nil {
		return nil, err
	}

	reply, err := hub.StartTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (t *tasksAPI) Status(ctx context.Context, id *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(t.ctx).Info("handling Status request", zap.String("id", id.Id))

	taskID, hubEth, err := util.ParseTaskID(id.Id)
	if err != nil {
		return nil, err
	}

	hubClient, err := t.getHubClientByEthAddr(ctx, hubEth)
	if err != nil {
		return nil, err
	}

	return hubClient.TaskStatus(ctx, &pb.ID{Id: taskID})
}

func (t *tasksAPI) Logs(req *pb.TaskLogsRequest, srv pb.TaskManagement_LogsServer) error {
	log.G(t.ctx).Info("handling Logs request", zap.Any("request", req))

	ctx := context.Background()
	_, hubEth, err := util.ParseTaskID(req.Id)
	if err != nil {
		return err
	}

	hubClient, err := t.getHubClientByEthAddr(ctx, hubEth)
	if err != nil {
		return err
	}

	logClient, err := hubClient.TaskLogs(ctx, req)
	if err != nil {
		return err
	}

	for {
		buffer, err := logClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		err = srv.Send(buffer)
		if err != nil {
			return err
		}
	}
}

func (t *tasksAPI) Stop(ctx context.Context, id *pb.ID) (*pb.Empty, error) {
	log.G(t.ctx).Info("handling Stop request", zap.String("id", id.Id))

	taskID, hubEth, err := util.ParseTaskID(id.Id)
	if err != nil {
		return nil, err
	}

	hubClient, err := t.getHubClientByEthAddr(ctx, hubEth)
	if err != nil {
		return nil, err
	}

	return hubClient.StopTask(ctx, &pb.ID{Id: taskID})
}

func (t *tasksAPI) getHubClientForDeal(ctx context.Context, id string) (pb.HubClient, error) {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return nil, err
	}

	dealInfo, err := t.bc.GetDealInfo(bigID)
	if err != nil {
		return nil, err
	}

	return t.getHubClientByEthAddr(ctx, dealInfo.GetSupplierID())
}

func (t *tasksAPI) getHubClientByEthAddr(ctx context.Context, eth string) (pb.HubClient, error) {
	resolve := &pb.ResolveRequest{EthAddr: eth}
	addrReply, err := t.locator.Resolve(ctx, resolve)
	if err != nil {
		return nil, err
	}

	cc, err := util.MakeGrpcClient(ctx, addrReply.IpAddr[0], t.creds)
	if err != nil {
		return nil, err
	}

	return pb.NewHubClient(cc), nil
}

func newTasksAPI(ctx context.Context, key *ecdsa.PrivateKey, conf Config, creds credentials.TransportCredentials) (pb.TaskManagementServer, error) {
	bcAPI, err := blockchain.NewBlockchainAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	cc, err := util.MakeGrpcClient(ctx, conf.LocatorEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	return &tasksAPI{
		ctx:     ctx,
		key:     key,
		bc:      bcAPI,
		locator: pb.NewLocatorClient(cc),
		creds:   creds,
	}, nil
}
