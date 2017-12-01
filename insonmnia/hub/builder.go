package hub

import (
	"crypto/ecdsa"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Builder struct {
	version       string
	ctx           context.Context
	ethKey        *ecdsa.PrivateKey
	ethAddr       string
	bcr           blockchain.Blockchainer
	market        pb.MarketClient
	locator       pb.LocatorClient
	creds         credentials.TransportCredentials
	rot           util.HitlessCertRotator
	cluster       Cluster
	clusterEvents <-chan ClusterEvent
}

func (b *Builder) WithContext(ctx context.Context) *Builder {
	b.ctx = ctx
	return b
}

func (b *Builder) WithBlockchain(bcr blockchain.Blockchainer) *Builder {
	b.bcr = bcr
	return b
}

func (b *Builder) WithMarket(m pb.MarketClient) *Builder {
	b.market = m
	return b
}

func (b *Builder) WithLocator(lc pb.LocatorClient) *Builder {
	b.locator = lc
	return b
}

func (b *Builder) WithPrivateKey(k *ecdsa.PrivateKey) *Builder {
	b.ethKey = k
	b.ethAddr = util.PubKeyToAddr(k.PublicKey)
	return b
}

func (b *Builder) WithVersion(v string) *Builder {
	b.version = v
	return b
}

func (b *Builder) WithCreds(creds credentials.TransportCredentials) *Builder {
	b.creds = creds
	return b
}

func (b *Builder) WithCertRotator(rot util.HitlessCertRotator) *Builder {
	b.rot = rot
	return b
}

func (b *Builder) WithCluster(cl Cluster, evt <-chan ClusterEvent) *Builder {
	b.cluster = cl
	b.clusterEvents = evt
	return b
}

func (b *Builder) Build(cfg *Config) (*Hub, error) {
	if b.ctx == nil {
		b.ctx = context.Background()
	}

	if b.ethKey == nil {
		return nil, errors.New("cannot build Hub instance without private key")
	}

	var err error
	ctx, cancel := context.WithCancel(b.ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	ip := cfg.EndpointIP()
	clientPort, err := util.ParseEndpointPort(cfg.Cluster.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "error during parsing client endpoint")
	}
	grpcEndpointAddr := ip + ":" + clientPort

	var gate *gateway.Gateway
	var portPool *gateway.PortPool
	if cfg.GatewayConfig != nil {
		gate, err = gateway.NewGateway(ctx)
		if err != nil {
			return nil, err
		}

		if len(cfg.GatewayConfig.Ports) != 2 {
			return nil, errors.New("gateway ports must be a range of two values")
		}

		portRangeFrom := cfg.GatewayConfig.Ports[0]
		portRangeSize := cfg.GatewayConfig.Ports[1] - portRangeFrom
		portPool = gateway.NewPortPool(portRangeFrom, portRangeSize)
	}

	ethWrapper, err := NewETH(ctx, b.ethKey, b.bcr)
	if err != nil {
		return nil, err
	}

	if b.locator == nil {
		conn, err := util.MakeGrpcClient(b.ctx, cfg.Locator.Address, b.creds, grpc.WithTimeout(5*time.Second))
		if err != nil {
			return nil, err
		}

		b.locator = pb.NewLocatorClient(conn)
	}

	if b.cluster == nil {
		b.cluster, b.clusterEvents, err = NewCluster(ctx, &cfg.Cluster, b.creds)
		if err != nil {
			return nil, err
		}
	}

	if os.Getenv("GRPC_INSECURE") != "" {
		b.rot = nil
		b.creds = nil
	}

	acl := NewACLStorage()
	if b.creds != nil {
		acl.Insert(b.ethAddr)
	}

	h := &Hub{
		cfg:              cfg,
		ctx:              ctx,
		cancel:           cancel,
		gateway:          gate,
		portPool:         portPool,
		externalGrpc:     nil,
		grpcEndpointAddr: grpcEndpointAddr,

		ethKey:  b.ethKey,
		ethAddr: b.ethAddr,
		version: b.version,

		locatorPeriod: time.Second * time.Duration(cfg.Locator.Period),
		locatorClient: b.locator,

		eth:    ethWrapper,
		market: b.market,

		tasks:            make(map[string]*TaskInfo),
		miners:           make(map[string]*MinerCtx),
		associatedHubs:   make(map[string]struct{}),
		deviceProperties: make(map[string]DeviceProperties),
		slots:            make(map[string]*structs.Slot),
		acl:              acl,

		certRotator: b.rot,
		creds:       b.creds,

		cluster:       b.cluster,
		clusterEvents: b.clusterEvents,
	}

	grpcServer := util.MakeGrpcServer(h.creds, grpc.UnaryInterceptor(h.onRequest))
	h.externalGrpc = grpcServer

	pb.RegisterHubServer(grpcServer, h)
	return h, nil
}

func NewBuilder() *Builder {
	return &Builder{}
}
