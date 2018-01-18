package node

import (
	"io"

	"strconv"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
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
	// has hubID, can perform direct request
	if req.GetHubID() != "" {
		log.G(t.ctx).Info("has HubAddr, performing direct request")
		hubClient, cc, err := getHubClientByEthAddr(ctx, t.remotes, req.GetHubID())
		if err != nil {
			return nil, err
		}
		defer cc.Close()

		return hubClient.TaskList(ctx, &pb.Empty{})
	}

	myAddr := util.PubKeyToAddr(t.remotes.key.PublicKey)
	dealIDs, err := t.remotes.eth.GetDeals(myAddr.Hex())
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
		t.getSupplierTasks(ctx, tasks, deal)
	}

	return &pb.TaskListReply{Info: tasks}, nil
}

func (t *tasksAPI) getSupplierTasks(ctx context.Context, tasks map[string]*pb.TaskListReply_TaskInfo, deal *pb.Deal) {
	hub, cc, err := getHubClientByEthAddr(ctx, t.remotes, deal.GetSupplierID())
	if err != nil {
		log.G(t.ctx).Error("cannot resolve hub address",
			zap.String("hub_eth", deal.GetSupplierID()),
			zap.Error(err))
		return
	}
	defer cc.Close()

	taskList, err := hub.TaskList(ctx, &pb.Empty{})
	if err != nil {
		log.G(t.ctx).Error("cannot retrieve tasks from the hub", zap.Error(err))
		return
	}

	for _, v := range taskList.GetInfo() {
		tasks[deal.GetSupplierID()] = v
	}
}

func (t *tasksAPI) Start(ctx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	hub, cc, err := getHubClientForDeal(ctx, t.remotes, req.Deal.GetId())
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	reply, err := hub.StartTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (t *tasksAPI) Status(ctx context.Context, id *pb.TaskID) (*pb.TaskStatusReply, error) {
	hubClient, cc, err := getHubClientByEthAddr(ctx, t.remotes, id.HubAddr)
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	return hubClient.TaskStatus(ctx, &pb.ID{Id: id.Id})
}

func (t *tasksAPI) Logs(req *pb.TaskLogsRequest, srv pb.TaskManagement_LogsServer) error {
	log.G(t.ctx).Info("handling Logs request", zap.Any("request", req))

	hubClient, cc, err := getHubClientByEthAddr(srv.Context(), t.remotes, req.HubAddr)
	if err != nil {
		return err
	}
	defer cc.Close()

	logClient, err := hubClient.TaskLogs(srv.Context(), req)
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
	hubClient, cc, err := getHubClientByEthAddr(ctx, t.remotes, id.HubAddr)
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	return hubClient.StopTask(ctx, &pb.ID{Id: id.Id})
}

func (t *tasksAPI) PushTask(clientStream pb.TaskManagement_PushTaskServer) error {
	meta, err := t.extractStreamMeta(clientStream)
	if err != nil {
		return err
	}

	log.G(t.ctx).Info("handling PushTask request", zap.String("deal_id", meta.dealID))

	hub, cc, err := getHubClientForDeal(meta.ctx, t.remotes, meta.dealID)
	if err != nil {
		return err
	}
	defer cc.Close()

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
					log.G(t.ctx).Debug("recieved last push chunk")
					clientCompleted = true
				} else {
					log.G(t.ctx).Debug("recieved push error", zap.Error(err))
					return err
				}
			}

			if chunk == nil {
				log.G(t.ctx).Debug("closing hub stream")
				if err := hubStream.CloseSend(); err != nil {
					return err
				}
			} else {
				bytesRemaining = len(chunk.Chunk)
				if err := hubStream.Send(chunk); err != nil {
					log.G(t.ctx).Debug("failed to send chunk to hub", zap.Error(err))
					return err
				}
				log.G(t.ctx).Debug("sent chunk to hub")
			}
		}

		for {
			progress, err := hubStream.Recv()
			if err != nil {
				if err == io.EOF {
					log.G(t.ctx).Debug("received last chunk from hub")
					if bytesCommitted == meta.fileSize {
						clientStream.SetTrailer(hubStream.Trailer())
						return nil
					} else {
						log.G(t.ctx).Debug("hub closed its stream without committing all bytes")
						return status.Errorf(codes.Aborted, "hub closed its stream without committing all bytes")
					}
				} else {
					log.G(t.ctx).Debug("received error from hub", zap.Error(err))
					return err
				}
			}

			bytesCommitted += progress.Size
			bytesRemaining -= int(progress.Size)

			if err := clientStream.Send(progress); err != nil {
				log.G(t.ctx).Debug("failed to send chunk back to cli", zap.Error(err))
				return err
			}

			if bytesRemaining == 0 {
				break
			}
		}
	}
}

func (t *tasksAPI) PullTask(req *pb.PullTaskRequest, srv pb.TaskManagement_PullTaskServer) error {
	ctx := context.Background()
	hub, cc, err := getHubClientForDeal(ctx, t.remotes, req.GetDealId())
	if err != nil {
		return err
	}
	defer cc.Close()

	pullClient, err := hub.PullTask(ctx, req)
	if err != nil {
		return err
	}

	header, err := pullClient.Header()
	if err != nil {
		return nil
	}

	err = srv.SetHeader(header)
	if err != nil {
		return nil
	}

	for {
		buffer, err := pullClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if buffer != nil {
			err = srv.Send(buffer)
			if err != nil {
				return err
			}
		}
	}
}

func getHubClientForDeal(ctx context.Context, rm *remoteOptions, id string) (pb.HubClient, io.Closer, error) {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return nil, nil, err
	}

	dealInfo, err := rm.eth.GetDealInfo(bigID)
	if err != nil {
		return nil, nil, err
	}

	return getHubClientByEthAddr(ctx, rm, dealInfo.GetSupplierID())
}

func getHubClientByEthAddr(ctx context.Context, rm *remoteOptions, eth string) (pb.HubClient, io.Closer, error) {
	resolve := &pb.ResolveRequest{EthAddr: eth}
	addrReply, err := rm.locator.Resolve(ctx, resolve)
	if err != nil {
		return nil, nil, err
	}

	// Maybe blocking connection required?
	cc, err := xgrpc.NewClient(ctx, addrReply.Endpoints[0], rm.creds)
	if err != nil {
		return nil, nil, err
	}

	return pb.NewHubClient(cc), cc, nil
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
