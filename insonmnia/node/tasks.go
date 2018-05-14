package node

import (
	"fmt"
	"io"
	"strconv"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
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
	if req.GetDealID() == nil || req.GetDealID().IsZero() {
		return nil, errors.New("deal ID is required for listing tasks")
	}
	log.G(t.ctx).Info("dealID is provided in request, performing direct request",
		zap.String("dealID", req.GetDealID().Unwrap().String()))

	dealID := req.GetDealID().Unwrap().String()
	hubClient, cc, err := t.remotes.getHubClientForDeal(ctx, dealID)
	if err != nil {
		return nil, err
	}
	defer cc.Close()
	deal, err := hubClient.GetDealInfo(ctx, &pb.ID{Id: req.GetDealID().Unwrap().String()})
	if err != nil {
		return nil, fmt.Errorf("failed to get deal info for deal %s: %s", dealID, err)
	}
	reply := &pb.TaskListReply{
		Info: map[string]*pb.TaskStatusReply{},
	}
	for id, task := range deal.Completed.Statuses {
		reply.Info[id] = task
	}
	for id, task := range deal.Running.Statuses {
		reply.Info[id] = task
	}
	return reply, nil
}

func (t *tasksAPI) getSupplierTasks(ctx context.Context, tasks map[string]*pb.TaskStatusReply, deal *pb.Deal) {
	dealID := deal.GetSupplierID().Unwrap().Hex()
	hub, cc, err := t.remotes.getHubClientByEthAddr(ctx, dealID)
	if err != nil {
		log.G(t.ctx).Error("cannot resolve worker address",
			zap.String("hub_eth", dealID),
			zap.Error(err))
		return
	}
	defer cc.Close()

	taskList, err := hub.Tasks(ctx, &pb.Empty{})
	if err != nil {
		log.G(t.ctx).Error("failed to retrieve tasks from the worker", zap.Error(err))
		return
	}

	for _, v := range taskList.GetInfo() {
		tasks[deal.GetSupplierID().Unwrap().Hex()] = v
	}
}

func (t *tasksAPI) Start(ctx context.Context, req *pb.StartTaskRequest) (*pb.StartTaskReply, error) {
	dealID := req.Deal.GetId().Unwrap().String()
	hub, cc, err := t.remotes.getHubClientForDeal(ctx, dealID)
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	reply, err := hub.StartTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start task on worker: %s", err)
	}

	return reply, nil
}

func (t *tasksAPI) JoinNetwork(ctx context.Context, request *pb.JoinNetworkRequest) (*pb.NetworkSpec, error) {
	dealID := request.GetTaskID().GetDealID().Unwrap().String()
	hub, cc, err := t.remotes.getHubClientForDeal(ctx, dealID)
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	reply, err := hub.JoinNetwork(ctx, &pb.HubJoinNetworkRequest{
		TaskID:    request.TaskID.Id,
		NetworkID: request.NetworkID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to join network on worker: %s", err)
	}

	return reply, nil
}

func (t *tasksAPI) Status(ctx context.Context, id *pb.TaskID) (*pb.TaskStatusReply, error) {
	hubClient, cc, err := t.remotes.getHubClientForDeal(ctx, id.GetDealID().Unwrap().String())
	if err != nil {
		return nil, err
	}
	defer cc.Close()

	return hubClient.TaskStatus(ctx, &pb.ID{Id: id.Id})
}

func (t *tasksAPI) Logs(req *pb.TaskLogsRequest, srv pb.TaskManagement_LogsServer) error {
	hubClient, cc, err := t.remotes.getHubClientForDeal(srv.Context(), req.GetDealID().Unwrap().String())
	if err != nil {
		return err
	}
	defer cc.Close()

	logClient, err := hubClient.TaskLogs(srv.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to fetch logs from worker: %s", err)
	}

	for {
		buffer, err := logClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("failure during receiving logs from worker: %s", err)
		}

		err = srv.Send(buffer)
		if err != nil {
			return fmt.Errorf("failed to send log chunk request to worker: %s", err)
		}
	}
}

func (t *tasksAPI) Stop(ctx context.Context, id *pb.TaskID) (*pb.Empty, error) {
	hubClient, cc, err := t.remotes.getHubClientForDeal(ctx, id.GetDealID().Unwrap().String())
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

	hub, cc, err := t.remotes.getHubClientForDeal(meta.ctx, meta.dealID)
	if err != nil {
		return err
	}
	defer cc.Close()

	hubStream, err := hub.PushTask(meta.ctx)
	if err != nil {
		return fmt.Errorf("failed to start task push server on worker: %s", err)
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
					return fmt.Errorf("failed to receive image chunk from client: %s", err)
				}
			}

			if chunk == nil {
				log.G(t.ctx).Debug("closing worker stream")
				if err := hubStream.CloseSend(); err != nil {
					return fmt.Errorf("failed to send closing frame to worker: %s", err)
				}
			} else {
				bytesRemaining = len(chunk.Chunk)
				if err := hubStream.Send(chunk); err != nil {
					log.G(t.ctx).Debug("failed to send chunk to worker", zap.Error(err))
					return fmt.Errorf("failed to send chunk to worker: %s", err)
				}
				log.G(t.ctx).Debug("sent chunk to worker")
			}
		}

		for {
			progress, err := hubStream.Recv()
			if err != nil {
				if err == io.EOF {
					log.G(t.ctx).Debug("received last chunk from worker")
					if bytesCommitted == meta.fileSize {
						clientStream.SetTrailer(hubStream.Trailer())
						return nil
					} else {
						log.G(t.ctx).Debug("worker closed its stream without committing all bytes")
						return status.Errorf(codes.Aborted, "worker closed its stream without committing all bytes")
					}
				} else {
					log.G(t.ctx).Debug("received error from worker", zap.Error(err))
					return fmt.Errorf("failed to receive meta info from worker: %s", err)
				}
			}

			bytesCommitted += progress.Size
			bytesRemaining -= int(progress.Size)

			if err := clientStream.Send(progress); err != nil {
				log.G(t.ctx).Debug("failed to send meta to client", zap.Error(err))
				return fmt.Errorf("failed to send meta to client: %s", err)
			}

			if bytesRemaining == 0 {
				break
			}
		}
	}
}

func (t *tasksAPI) PullTask(req *pb.PullTaskRequest, srv pb.TaskManagement_PullTaskServer) error {
	ctx := context.Background()
	hub, cc, err := t.remotes.getHubClientForDeal(ctx, req.GetDealId())
	if err != nil {
		return err
	}
	defer cc.Close()

	pullClient, err := hub.PullTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start task pull server on worker: %s", err)
	}

	header, err := pullClient.Header()
	if err != nil {
		return fmt.Errorf("failed to receive meta from worker: %s", err)
	}

	err = srv.SetHeader(header)
	if err != nil {
		return fmt.Errorf("failed to set meta for client: %s", err)
	}

	for {
		buffer, err := pullClient.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to receive chunk from worker: %s", err)
		}

		if buffer != nil {
			err = srv.Send(buffer)
			if err != nil {
				return fmt.Errorf("failed to send meta to client: %s", err)
			}
		}
	}
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
