package node

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type workerAPI struct {
	pb.WorkerManagementServer
	remotes *remoteOptions
	ctx     context.Context
}

func (h *workerAPI) getWorkerEthAddr(ctx context.Context) (common.Address, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return crypto.PubkeyToAddress(h.remotes.key.PublicKey), nil
	}
	ctxAddrs, ok := md["x_worker_eth_addr"]
	if !ok || len(ctxAddrs) == 0 {
		return crypto.PubkeyToAddress(h.remotes.key.PublicKey), nil
	}
	return util.HexToAddress(ctxAddrs[0])
}

func getWorkerNetAddr(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	ctxAddrs, ok := md["x_worker_net_addr"]
	if !ok || len(ctxAddrs) == 0 {
		return ""
	}
	return ctxAddrs[0]
}

func (h *workerAPI) getClient(ctx context.Context) (pb.WorkerManagementClient, io.Closer, error) {
	ethAddr, err := h.getWorkerEthAddr(ctx)
	if err != nil {
		return nil, nil, err
	}

	netAddr := getWorkerNetAddr(ctx)

	log.S(h.ctx).Debugf("connecting to worker on %s@%s", ethAddr.Hex(), netAddr)
	return h.remotes.workerCreator(ethAddr, netAddr)
}

func (h *workerAPI) intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	methodName := util.ExtractMethod(info.FullMethod)

	log.S(h.ctx).Infof("handling %s request", methodName)

	ctx = util.ForwardMetadata(ctx)
	if !strings.HasPrefix(info.FullMethod, "/sonm.WorkerManagement") {
		return handler(ctx, req)
	}

	cli, cc, err := h.getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to worker: %s", err)
	}
	defer cc.Close()

	var (
		t        = reflect.ValueOf(cli)
		method   = t.MethodByName(methodName)
		inValues = []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
		values   = method.Call(inValues)
	)
	if !values[1].IsNil() {
		err = values[1].Interface().(error)
	}

	return values[0].Interface(), err
}

func newWorkerAPI(opts *remoteOptions) pb.WorkerManagementServer {
	return &workerAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}
}
