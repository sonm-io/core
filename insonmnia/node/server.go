package node

import (
	"net"

	"github.com/jinzhu/configor"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
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
	// HubEndpoint is Hub's gRPC endpoint (not required)
	HubEndpoint() string
	// LogLevel return log verbosity
	LogLevel() int
	// ClientID returns EtherumID of Node Owner
	ClientID() string
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

func (y *yamlConfig) ClientID() string {
	// NOTE: just for testing on current iteration
	// key exchange will be implemented soon

	return "my-uniq-id"
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
	srv  *grpc.Server
}

// New creates new Local Node instance
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func New(ctx context.Context, c Config) (*Node, error) {
	lis, err := net.Listen("tcp", c.ListenAddress())
	if err != nil {
		return nil, err
	}

	srv := util.MakeGrpcServer()
	// register hub connection if hub addr is set
	if c.HubEndpoint() != "" {
		hub, err := newHubAPI(ctx, c)
		if err != nil {
			return nil, err
		}
		pb.RegisterHubManagementServer(srv, hub)
		log.G(ctx).Info("hub service registered", zap.String("endpt", c.HubEndpoint()))
	}

	market, err := newMarketAPI(ctx, c)
	if err != nil {
		return nil, err
	}

	pb.RegisterMarketServer(srv, market)
	log.G(ctx).Info("market service registered", zap.String("endpt", c.MarketEndpoint()))

	deals := newDealsAPI()
	pb.RegisterDealManagementServer(srv, deals)

	tasks := newTasksAPI()
	pb.RegisterTaskManagementServer(srv, tasks)

	return &Node{
		lis:  lis,
		conf: c,
		ctx:  ctx,
		srv:  srv,
	}, nil
}

// Serve binds gRPC services and start it
func (n *Node) Serve() error {
	return n.srv.Serve(n.lis)
}
