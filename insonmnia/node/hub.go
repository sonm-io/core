package node

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp"
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
	rendezvousEndpoints, err := h.remotes.conf.NPPConfig().Rendezvous.ConvertEndpoints()
	if err != nil {
		return nil, nil, err
	}

	relayEndpoints, err := h.remotes.conf.NPPConfig().Relay.ConvertEndpoints()
	if err != nil {
		return nil, nil, err
	}

	hubETH := crypto.PubkeyToAddress(h.remotes.key.PublicKey)

	dial, err := npp.NewDialer(h.ctx, npp.WithRendezvous(rendezvousEndpoints, h.remotes.creds), npp.WithRelayClient(relayEndpoints, hubETH))
	if err != nil {
		return nil, nil, err
	}

	addr := auth.NewAddrFromParts(hubETH, h.remotes.conf.HubEndpoint())
	conn, err := dial.Dial(addr)
	if err != nil {
		return nil, nil, err
	}

	cc, err := xgrpc.NewClient(h.ctx, "-", auth.NewWalletAuthenticator(h.remotes.creds, hubETH), xgrpc.WithConn(conn))
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
		return nil, errors.New("hub endpoint is not configured, please check Node settings")
	}

	cli, cc, err := h.getClient()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to hub at %s, please check Node settings: %s", h.remotes.conf.HubEndpoint(), err.Error())
	}
	defer cc.Close()

	mappedName, ok := hubToNodeMethods[methodName]
	if !ok {
		return nil, fmt.Errorf("unknwon management api method \"%s\"", methodName)
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
var hubToNodeMethods = map[string]string{
	"Status":        "Status",
	"Devices":       "Devices",
	"Tasks":         "Tasks",
	"AskPlans":      "AskPlans",
	"CreateAskPlan": "CreateAskPlan",
	"RemoveAskPlan": "RemoveAskPlan",
}

func newHubAPI(opts *remoteOptions) pb.HubManagementServer {
	return &hubAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}
}
