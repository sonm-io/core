package node

import (
	"net"

	"github.com/jinzhu/configor"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Config is LocalNode config
type Config interface {
	// ListenAddress is gRPC endpoint that Node binds to
	ListenAddress() string
	// MarketEndpoint is Marketplace gRPC endpoint
	MarketEndpoint() string
	// HubEndpoint is Hub's gRPC endpoint (maybe empty)
	HubEndpoint() string
	// LogLevel return log verbosity
	LogLevel() int
}

type nodeConfig struct {
	ListenAddr string `required:"true" yaml:"listen_addr"`
}

type marketConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type hubConfig struct {
	Endpoint string `required:"false" yaml:"endpoint"`
}

type logConfig struct {
	Level int `required:"true" default:"-1" yaml:"level"`
}

type yamlConfig struct {
	Node   nodeConfig   `required:"true" yaml:"node"`
	Market marketConfig `required:"true" yaml:"market"`
	Log    logConfig    `required:"true" yaml:"log"`
	Hub    *hubConfig   `required:"false" yaml:"hub"`
}

func (y *yamlConfig) ListenAddress() string {
	return y.Node.ListenAddr
}

func (y *yamlConfig) MarketEndpoint() string {
	return y.Market.Endpoint
}

func (y *yamlConfig) HubEndpoint() string {
	if y.Hub != nil {
		return y.Hub.Endpoint
	}
	return ""
}

func (y *yamlConfig) LogLevel() int {
	return y.Log.Level
}

// NewConfig loads localNode config from given .yaml file
func NewConfig(path string) (Config, error) {
	cfg := &yamlConfig{}

	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Node is LocalNode instance
type Node struct {
	ctx  context.Context
	conf Config
	lis  net.Listener
}

// New creates new Local Node instance
func New(ctx context.Context, c Config) (*Node, error) {
	lis, err := net.Listen("tcp", c.ListenAddress())
	if err != nil {
		return nil, err
	}

	return &Node{
		lis:  lis,
		conf: c,
		ctx:  ctx,
	}, nil
}

// Serve binds gRPC services and start it
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func (n *Node) Serve() error {
	srv := grpc.NewServer(
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()))

	// register hub connection if hub addr is set
	if n.conf.HubEndpoint() != "" {
		hub, err := newHubAPI(n.ctx, n.conf.HubEndpoint())
		if err != nil {
			return err
		}
		pb.RegisterHubManagementServer(srv, hub)
		log.G(n.ctx).Info("hub service registered", zap.String("endpt", n.conf.HubEndpoint()))
	}

	market, err := newMarketAPI(n.ctx, n.conf.MarketEndpoint())
	if err != nil {
		return err
	}

	pb.RegisterMarketServer(srv, market)
	log.G(n.ctx).Info("market service registered", zap.String("endpt", n.conf.MarketEndpoint()))

	deals := newDealsAPI()
	pb.RegisterDealManagementServer(srv, deals)

	tasks := newTasksAPI()
	pb.RegisterTaskManagementServer(srv, tasks)

	return srv.Serve(n.lis)
}
