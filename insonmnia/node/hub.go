package node

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type hubAPI struct {
	pb.HubManagementServer
	remotes *remoteOptions
	ctx     context.Context
}

func (h *hubAPI) getClient() (pb.HubClient, io.Closer, error) {
	cc, err := xgrpc.NewClient(h.ctx, h.remotes.conf.HubEndpoint(), h.remotes.creds,
		grpc.WithBlock(), grpc.WithTimeout(15*time.Second))
	if err != nil {
		return nil, nil, err
	}

	return pb.NewHubClient(cc), cc, nil
}

func (h *hubAPI) intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	methodName := util.ExtractMethod(info.FullMethod)

	log.S(h.ctx).Infof("handling %s request", methodName)

	ctx = util.ForwardMetadata(ctx)
	if !strings.HasPrefix(info.FullMethod, "/sonm.HubManagement") {
		return handler(ctx, req)
	}

	if h.remotes.conf.HubEndpoint() == "" {
		return nil, errHubEndpointIsNotSet
	}

	cli, cc, err := h.getClient()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to hub at %s, please check Node settings", h.remotes.conf.HubEndpoint())
	}
	defer cc.Close()

	var (
		t          = reflect.ValueOf(cli)
		mappedName = hubToNodeMethods[methodName]
		method     = t.MethodByName(mappedName)
		inValues   = []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
		values     = method.Call(inValues)
	)
	if !values[1].IsNil() {
		err = values[1].Interface().(error)
	}

	return values[0].Interface(), err
}

// we need this because of not all of the methods can be mapped one-to-one between Node and Hub
// The more simplest way to omit this mapping is to refactor Hub's proto definition
// (not the Node's one because of the Node API is publicly declared and must be changed as rare as possible).
var hubToNodeMethods = map[string]string{
	"Status":               "Status",
	"WorkersList":          "List",
	"WorkerStatus":         "Info",
	"GetRegisteredWorkers": "GetRegisteredWorkers",
	"RegisterWorker":       "RegisterWorker",
	"DeregisterWorker":     "DeregisterWorker",
	"DeviceList":           "Devices",
	"GetDeviceProperties":  "GetDeviceProperties",
	"SetDeviceProperties":  "SetDeviceProperties",
	"GetAskPlan":           "GetAskPlan",
	"GetAskPlans":          "Slots",
	"CreateAskPlan":        "InsertSlot",
	"RemoveAskPlan":        "RemoveSlot",
	"TaskList":             "TaskList",
	"TaskStatus":           "TaskStatus",
}

func newHubAPI(opts *remoteOptions) pb.HubManagementServer {
	return &hubAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}
}
