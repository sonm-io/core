package node

import (
	"io"

	"strconv"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type tasksAPI struct {
	ctx     context.Context
	remotes *remoteOptions
}

func (t *tasksAPI) List(ctx context.Context, req *pb.TaskListRequest) (*pb.TaskListReply, error) {
	log.G(t.ctx).Info("handling List request", zap.Any("request", req))

	// has hubID, can perform direct request
	if req.GetHubID() != "" {
		log.G(t.ctx).Info("has HubAddr, performing direct request")
		hubClient, err := t.getHubClientByEthAddr(ctx, req.GetHubID())
		if err != nil {
			return nil, err
		}

		return hubClient.TaskList(ctx, &pb.Empty{})
	}

	myAddr := util.PubKeyToAddr(t.remotes.key.PublicKey)
	dealIDs, err := t.remotes.eth.GetDeals(myAddr)
	if err != nil {
		return nil, err
	}

	log.G(t.ctx).Info("found some deals", zap.Int("count", len(dealIDs)))

	var activeDeals []*pb.Deal
	for _, id := range dealIDs {
		dealInfo, err := t.remotes.eth.GetDealInfo(id)
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

func (t *tasksAPI) Status(ctx context.Context, id *pb.TaskID) (*pb.TaskStatusReply, error) {
	log.G(t.ctx).Info("handling Status request", zap.String("id", id.Id))

	hubClient, err := t.getHubClientByEthAddr(ctx, id.HubAddr)
	if err != nil {
		return nil, err
	}

	return hubClient.TaskStatus(ctx, &pb.ID{Id: id.Id})
}

func (t *tasksAPI) Logs(req *pb.TaskLogsRequest, srv pb.TaskManagement_LogsServer) error {
	log.G(t.ctx).Info("handling Logs request", zap.Any("request", req))

	ctx := context.Background()
	hubClient, err := t.getHubClientByEthAddr(ctx, req.HubAddr)
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

func (t *tasksAPI) Stop(ctx context.Context, id *pb.TaskID) (*pb.Empty, error) {
	log.G(t.ctx).Info("handling Stop request", zap.String("id", id.Id))

	hubClient, err := t.getHubClientByEthAddr(ctx, id.HubAddr)
	if err != nil {
		return nil, err
	}

	return hubClient.StopTask(ctx, &pb.ID{Id: id.Id})
}

func (t *tasksAPI) PushTask(clientStream pb.TaskManagement_PushTaskServer) error {
	meta, err := t.extractStreamMeta(clientStream)
	if err != nil {
		return err
	}

	log.G(t.ctx).Info("handling PushTask request", zap.String("deal_id", meta.dealID))

	cc, err := util.MakeGrpcClient(meta.ctx, "[::]:10001", t.remotes.creds)
	hub := pb.NewHubClient(cc)

	// hub, err := t.getHubClientForDeal(meta.ctx, meta.dealID)
	if err != nil {
		return err
	}

	hubStream, err := hub.PushTask(meta.ctx)
	if err != nil {
		return err
	}

	bytesCommitted := int64(0)
	clientCompleted := false

	for {
		bytesRemaining := 0
		if !clientCompleted {
			chunk, err := clientStream.Recv()
			if err != nil {
				if err == io.EOF {
					clientCompleted = true
				} else {
					return err
				}
			}

			if chunk == nil {
				if err := hubStream.CloseSend(); err != nil {
					return err
				}
			} else {
				bytesRemaining = len(chunk.Chunk)
				if err := hubStream.Send(chunk); err != nil {
					return err
				}
			}
		}

		for {
			progress, err := hubStream.Recv()
			if err != nil {
				if err == io.EOF {
					if bytesCommitted == meta.fileSize {
						clientStream.SetTrailer(hubStream.Trailer())
						return nil
					} else {
						return status.Errorf(codes.Aborted, "miner closed its stream without committing all bytes")
					}
				} else {
					return err
				}
			}

			bytesCommitted += progress.Size
			bytesRemaining -= int(progress.Size)

			if err := clientStream.Send(progress); err != nil {
				return err
			}

			if bytesRemaining == 0 {
				break
			}
		}
	}

	return nil
}

func (t *tasksAPI) PullTask(req *pb.PullTaskRequest, srv pb.TaskManagement_PullTaskServer) error {
	log.G(t.ctx).Info("handling PullTask request",
		zap.String("deal_id", req.DealId),
		zap.String("task_id", req.TaskId),
		zap.String("name", req.Name))

	ctx := context.Background()

	cc, err := util.MakeGrpcClient(ctx, "[::]:10001", t.remotes.creds)
	hub := pb.NewHubClient(cc)

	// TODO(sshaman1101): get it back
	// hub, err := t.getHubClientForDeal(ctx, req.GetDealId())
	if err != nil {
		return err
	}

	pullClient, err := hub.PullTask(ctx, req)
	if err != nil {
		return err
	}

	for {
		buffer, err := pullClient.Recv()
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

	return nil
}

func (t *tasksAPI) getHubClientForDeal(ctx context.Context, id string) (pb.HubClient, error) {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return nil, err
	}

	dealInfo, err := t.remotes.eth.GetDealInfo(bigID)
	if err != nil {
		return nil, err
	}

	return t.getHubClientByEthAddr(ctx, dealInfo.GetSupplierID())
}

func (t *tasksAPI) getHubClientByEthAddr(ctx context.Context, eth string) (pb.HubClient, error) {
	resolve := &pb.ResolveRequest{EthAddr: eth}
	addrReply, err := t.remotes.locator.Resolve(ctx, resolve)
	if err != nil {
		return nil, err
	}

	// maybe blocking connection required?
	cc, err := util.MakeGrpcClient(ctx, addrReply.IpAddr[0], t.remotes.creds)
	if err != nil {
		return nil, err
	}

	return pb.NewHubClient(cc), nil
}

type streamMeta struct {
	ctx      context.Context
	dealID   string
	fileSize int64
}

func (t *tasksAPI) extractStreamMeta(clientStream pb.TaskManagement_PushTaskServer) (*streamMeta, error) {
	md, ok := metadata.FromIncomingContext(clientStream.Context())
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "metadata required")
	}

	dealIDs, ok := md["deal"]
	if !ok || len(dealIDs) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "`%s` required", "deal")
	}

	sizes, ok := md["size"]
	if !ok || len(sizes) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "`%s` required", "size")
	}

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		"deal": dealIDs[0],
		"size": sizes[0],
	}))

	v, _ := strconv.ParseInt(sizes[0], 10, 64)

	return &streamMeta{
		ctx:      ctx,
		dealID:   dealIDs[0],
		fileSize: v,
	}, nil
}

func newTasksAPI(opts *remoteOptions) (pb.TaskManagementServer, error) {
	return &tasksAPI{
		ctx:     opts.ctx,
		remotes: opts,
	}, nil
}
