package node

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type interceptedAPI struct {
	sonm.WorkerServer
	sonm.WorkerManagementServer
	sonm.DWHServer
	remotes *remoteOptions
	log     *zap.SugaredLogger
}

func (m *interceptedAPI) getWorkerAddr(ctx context.Context) (*auth.Addr, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return auth.NewETHAddr(crypto.PubkeyToAddress(m.remotes.key.PublicKey)), nil
	}
	ctxAddrs, ok := md[util.WorkerAddressHeader]
	if !ok {
		return auth.NewETHAddr(crypto.PubkeyToAddress(m.remotes.key.PublicKey)), nil
	}
	if len(ctxAddrs) != 1 {
		return nil, fmt.Errorf("worker address key in metadata has %d headers (exactly one required)", len(ctxAddrs))
	}
	return auth.ParseAddr(ctxAddrs[0])
}

func (m *interceptedAPI) getWorkerManagementClient(ctx context.Context) (sonm.WorkerManagementClient, io.Closer, error) {
	addr, err := m.getWorkerAddr(ctx)
	if err != nil {
		return nil, nil, err
	}

	m.log.Debugf("connecting to worker on %s", addr.String())
	return m.remotes.workerCreator(ctx, addr)
}

func (m *interceptedAPI) getWorkerClient(ctx context.Context) (sonm.WorkerClient, io.Closer, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil, status.Errorf(codes.InvalidArgument, "metadata required")
	}

	dealIDs, ok := md["deal"]
	if !ok || len(dealIDs) != 1 {
		return nil, nil, status.Errorf(codes.InvalidArgument, "deal field is required and should be unique")
	}

	worker, cc, err := m.remotes.getWorkerClientForDeal(ctx, dealIDs[0])
	if err != nil {
		return nil, nil, err
	}

	return worker, cc, nil
}

func (m *interceptedAPI) intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var cli interface{}
	var closer io.Closer
	var err error

	serverName := strings.Split(info.FullMethod, "/")[1]
	switch serverName {
	case "sonm.Worker":
		ctx = util.ForwardMetadata(ctx)
		cli, closer, err = m.getWorkerClient(ctx)
	case "sonm.WorkerManagement":
		ctx = util.ForwardMetadata(ctx)
		cli, closer, err = m.getWorkerManagementClient(ctx)
	case "sonm.DWH":
		cli = m.remotes.dwh
	default:
		return handler(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	if closer != nil {
		defer closer.Close()
	}

	t := reflect.ValueOf(cli)
	method := t.MethodByName(xgrpc.ParseMethodInfo(info.FullMethod).Method)
	inValues := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)}
	values := method.Call(inValues)

	if !values[1].IsNil() {
		err = values[1].Interface().(error)
	}

	return values[0].Interface(), err
}

func callMethod(cli interface{}, methodStr string, args ...interface{}) []reflect.Value {
	value, ok := cli.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(cli)
	}
	inValues := []reflect.Value{}
	for _, arg := range args {
		valueArg, ok := arg.(reflect.Value)
		if !ok {
			valueArg = reflect.ValueOf(arg)
		}
		inValues = append(inValues, valueArg)
	}
	method := value.MethodByName(methodStr)
	return method.Call(inValues)
}

func callErrMethod(cli interface{}, methodStr string, args ...interface{}) error {
	retValues := callMethod(cli, methodStr, args...)
	if len(retValues) != 1 {
		panic(fmt.Sprintf("failed to call method %s, it has %d return values, required 1", methodStr, len(retValues)))
	}
	if !retValues[0].IsNil() {
		return retValues[0].Interface().(error)
	}
	return nil
}

func callBinMethod(cli interface{}, methodStr string, args ...interface{}) (reflect.Value, error) {
	retValues := callMethod(cli, methodStr, args...)
	if len(retValues) != 2 {
		panic(fmt.Sprintf("failed to call method %s, it has %d return values, required 2", methodStr, len(retValues)))
	}
	var err error
	if !retValues[1].IsNil() {
		err = retValues[1].Interface().(error)
	}
	return retValues[0], err
}

func newMethodArgValue(cli interface{}, methodStr string, position int) reflect.Value {
	value, ok := cli.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(cli)
	}
	method := value.MethodByName(methodStr)
	return reflect.New(method.Type().In(position).Elem())
}

func (m *interceptedAPI) streamIntercept(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	//TODO: deduplicate
	var cli interface{}
	var closer io.Closer
	var err error
	var ctx context.Context

	serverName := strings.Split(info.FullMethod, "/")[1]
	switch serverName {
	case "sonm.Worker":
		ctx = util.ForwardMetadata(ss.Context())
		cli, closer, err = m.getWorkerClient(ctx)
	case "sonm.WorkerManagement":
		ctx = util.ForwardMetadata(ss.Context())
		cli, closer, err = m.getWorkerClient(ctx)
	default:
		return handler(srv, ss)
	}
	if err != nil {
		return err
	}

	defer closer.Close()

	args := []interface{}{ctx}
	methodName := xgrpc.ParseMethodInfo(info.FullMethod).Method

	if !info.IsClientStream {
		// first position is always context, second is request
		request := newMethodArgValue(cli, methodName, 1)
		ss.RecvMsg(request.Interface())
		args = append(args, request)
	}
	streamCli, err := callBinMethod(cli, methodName, args...)
	if err != nil {
		return err
	}

	wg := errgroup.Group{}
	wg.Go(func() error {
		if !info.IsClientStream {
			return nil
		}
		for {
			sendVar := newMethodArgValue(streamCli, "Send", 0)
			if err := ss.RecvMsg(sendVar.Interface()); err != nil {
				//TODO: CloseAndRecv for 3d case
				if err == io.EOF {
					return streamCli.Interface().(grpc.ClientStream).CloseSend()
				}
				return err
			}
			if err := streamCli.Interface().(grpc.ClientStream).SendMsg(sendVar.Interface()); err != nil {
				return err
			}
		}
	})

	wg.Go(func() error {
		if !info.IsServerStream {
			return nil
		}
		headers, err := streamCli.Interface().(grpc.ClientStream).Header()
		if err != nil {
			return err
		}
		if err := ss.SendHeader(headers); err != nil {
			return fmt.Errorf("failed to send metadata back to client: %s", err)
		}
		for {
			progress, err := callBinMethod(streamCli, "Recv")
			if err == io.EOF {
				trailer := streamCli.Interface().(grpc.ClientStream).Trailer()
				ss.SetTrailer(trailer)
				return nil
			}
			if err != nil {
				return err
			}
			if err := ss.SendMsg(progress.Interface()); err != nil {
				return err
			}
		}

	})
	return wg.Wait()
}

func newInterceptedAPI(opts *remoteOptions) *interceptedAPI {
	return &interceptedAPI{
		remotes: opts,
		log:     opts.log,
	}
}
