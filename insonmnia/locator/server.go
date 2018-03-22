package locator

import (
	"crypto/ecdsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/ethereum/go-ethereum/common"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/auth"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var errNodeNotFound = errors.New("record with given Eth address cannot be found")

type record struct {
	EthAddr         common.Address
	ClientEndpoints []string
	WorkerEndpoints []string
	TS              time.Time
}

type Locator struct {
	conf                *Config
	ctx                 context.Context
	grpc                *grpc.Server
	certRotator         util.HitlessCertRotator
	creds               credentials.TransportCredentials
	onlyPublicClientIPs bool
	storage             store.Store
}

func (l *Locator) Announce(ctx context.Context, req *pb.AnnounceRequest) (*pb.Empty, error) {
	ethAddr, err := l.extractEthAddr(ctx)
	if err != nil {
		return nil, err
	}

	clientEndpoints, err := l.filterEndpoints(ethAddr, req.ClientEndpoints, l.onlyPublicClientIPs)
	if err != nil {
		return nil, errors.Wrap(err, "invalid client endpoints")
	}

	//workerEndpoints, err := l.filterEndpoints(ethAddr, req.WorkerEndpoints, false)
	//if err != nil {
	//	return nil, errors.Wrap(err, "invalid worker endpoints")
	//}

	log.G(l.ctx).Info("handling Announce request",
		zap.Stringer("eth", ethAddr),
		zap.Strings("client_endpoints", clientEndpoints),
	)

	err = l.put(&record{
		EthAddr:         ethAddr,
		ClientEndpoints: clientEndpoints,
	})

	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (l *Locator) filterEndpoints(ethAddr common.Address, endpoints []string, onlyPublic bool) ([]string, error) {
	var okEndpoints, skippedEndpoints []string
	for _, endpoint := range endpoints {
		strIP, _, err := net.SplitHostPort(endpoint)
		if err != nil {
			skippedEndpoints = append(skippedEndpoints, endpoint)
			continue
		}

		if onlyPublic {
			if ip := net.ParseIP(strIP); ip != nil && !util.IsPrivateIP(ip) {
				okEndpoints = append(okEndpoints, endpoint)
			} else {
				skippedEndpoints = append(skippedEndpoints, endpoint)
			}
		} else {
			okEndpoints = append(okEndpoints, endpoint)
		}
	}

	if len(skippedEndpoints) > 0 {
		log.G(l.ctx).Info("skipped some announced endpoints (only public IPs mode is on)",
			zap.Stringer("eth", ethAddr),
			zap.Strings("skipped_ips", skippedEndpoints))
	}

	if len(okEndpoints) < 1 {
		return nil, errors.New("no white IPs provided")
	}

	return okEndpoints, nil
}

func (l *Locator) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveReply, error) {
	log.G(l.ctx).Info("handling Resolve request", zap.String("eth", req.EthAddr))

	if !common.IsHexAddress(req.EthAddr) {
		return nil, fmt.Errorf("invalid ethaddress %s", req.EthAddr)
	}

	rec, err := l.get(common.HexToAddress(req.EthAddr))
	if err != nil {
		return nil, err
	}

	var endpoints []string
	switch req.EndpointType {
	case pb.ResolveRequest_CLIENT:
		endpoints = rec.ClientEndpoints
	case pb.ResolveRequest_WORKER:
		endpoints = rec.WorkerEndpoints
	case pb.ResolveRequest_ANY:
		endpoints = append(rec.ClientEndpoints, rec.WorkerEndpoints...)
	default:
		return nil, fmt.Errorf("unknown endpoint type: %d", req.EndpointType)
	}

	return &pb.ResolveReply{Endpoints: endpoints}, nil
}

func (l *Locator) Serve() error {
	lis, err := net.Listen("tcp", l.conf.ListenAddr)
	if err != nil {
		return err
	}

	return l.grpc.Serve(lis)
}

func (l *Locator) extractEthAddr(ctx context.Context) (common.Address, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return common.Address{}, status.Error(codes.DataLoss, "failed to get peer from ctx")
	}

	switch authInfo := pr.AuthInfo.(type) {
	case auth.EthAuthInfo:
		return authInfo.Wallet, nil
	default:
		return common.Address{}, status.Error(codes.Unauthenticated, "wrong AuthInfo type")
	}
}

func (l *Locator) put(rec *record) error {
	rec.TS = time.Now()
	key := rec.EthAddr.Hex()
	value, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	return l.storage.Put(key, value, nil)
}

func (l *Locator) get(ethAddr common.Address) (*record, error) {
	key := ethAddr.Hex()

	pair, err := l.storage.Get(key)
	if err != nil {
		log.G(l.ctx).Debug("record not found", zap.String("key", key))
		return nil, errNodeNotFound
	}

	rec := &record{}
	if err = json.Unmarshal(pair.Value, rec); err != nil {
		return nil, err
	}

	notBefore := time.Now().Add(-1 * l.conf.NodeTTL)
	if rec.TS.Before(notBefore) {
		log.G(l.ctx).Debug("record is expired", zap.String("key", key))
		l.storage.Delete(pair.Key)
		return nil, errNodeNotFound
	}

	return rec, nil
}

func initStorage(ctx context.Context, conf storeConfig) (store.Store, error) {
	consul.Register()
	boltdb.Register()

	log.G(ctx).Info("creating store", zap.Any("store", conf))

	endpoints := []string{conf.Endpoint}
	backend := store.Backend(conf.Type)

	config := store.Config{
		TLS:    nil,
		Bucket: conf.Bucket,
	}

	storage, err := libkv.NewStore(backend, endpoints, &config)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func NewLocator(ctx context.Context, conf *Config, key *ecdsa.PrivateKey) (l *Locator, err error) {
	if key == nil {
		return nil, errors.Wrap(err, "private key should be provided")
	}

	l = &Locator{
		conf:                conf,
		ctx:                 ctx,
		onlyPublicClientIPs: conf.OnlyPublicClientIPs,
	}

	var TLSConfig *tls.Config
	l.certRotator, TLSConfig, err = util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}

	s, err := initStorage(ctx, conf.Store)
	if err != nil {
		return nil, err
	}

	l.storage = s
	l.creds = util.NewTLS(TLSConfig)
	l.grpc = xgrpc.NewServer(log.GetLogger(l.ctx),
		xgrpc.Credentials(l.creds),
		xgrpc.DefaultTraceInterceptor(),
	)

	pb.RegisterLocatorServer(l.grpc, l)
	grpc_prometheus.Register(l.grpc)

	return l, nil
}
