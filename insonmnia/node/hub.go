package node

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type hubAPI struct {
	pb.WorkerManagementServer
	remotes *remoteOptions
	ctx     context.Context
}

func (h *hubAPI) getClient() (pb.WorkerManagementClient, io.Closer, error) {
	ethAddr := crypto.PubkeyToAddress(h.remotes.key.PublicKey)
	netAddr := h.remotes.conf.Hub.Endpoint

	return h.remotes.hubCreator(ethAddr, netAddr)
}

func (h *hubAPI) intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	methodName := util.ExtractMethod(info.FullMethod)

	log.S(h.ctx).Infof("handling %s request", methodName)

	ctx = util.ForwardMetadata(ctx)
	if !strings.HasPrefix(info.FullMethod, "/sonm.WorkerManagement") {
		return handler(ctx, req)
	}

	if h.remotes.conf.Hub.Endpoint == "" {
		return nil, errors.New("hub endpoint is not configured, please check Node settings")
	}

	cli, cc, err := h.getClient()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to hub at %s, please check Node settings: %s", h.remotes.conf.Hub.Endpoint, err.Error())
	}
	defer cc.Close()

	mappedName, ok := hubToNodeMethods[methodName]
	if !ok {
		return nil, fmt.Errorf("unknwon management API method \"%s\"", methodName)
	}

	var (
		t        = reflect.ValueOf(cli)
		method   = t.MethodByName(mappedName)
		inValues = []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
		values   = method.Call(inValues)
	)
	if !values[1].IsNil() {
		err = values[1].Interface().(error)
	}

	return values[0].Interface(), err
}

// we need this because of not all of the methods can be mapped one-to-one between Node and Hub
// The more simplest way to omit this mapping is to refactor Hub's proto definition
// (not the Node's one because of the Node API is publicly declared and must be changed as rare as possible).
//
// Deprecated: methods maps one-to-one, so this mapping become meaningless.
var hubToNodeMethods = map[string]string{
	"Status":        "Status",
	"Devices":       "Devices",
	"Tasks":         "Tasks",
	"AskPlans":      "AskPlans",
	"CreateAskPlan": "CreateAskPlan",
	"RemoveAskPlan": "RemoveAskPlan",
}

func newHubAPI(opts *remoteOptions) pb.WorkerManagementServer {
	return &hubAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}
}
