package node

import (
	"net"

	"crypto/ecdsa"

	"github.com/jinzhu/configor"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/accounts"
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
	// LocatorEndpoint is Locator service gRPC endpoint
	LocatorEndpoint() string
	// LogLevel return log verbosity
	LogLevel() int
	// KeyStorager included into config because of
	// Node instance must know how to open the keystore
	accounts.KeyStorager
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

type locatorConfig struct {
	Endpoint string `required:"true" default:"" yaml:"endpoint"`
}

type keysConfig struct {
	Keystore   string `required:"false" default:"" yaml:"key_store"`
	Passphrase string `required:"false" default:"" yaml:"pass_phrase"`
}

type yamlConfig struct {
	Node    nodeConfig    `required:"true" yaml:"node"`
	Market  marketConfig  `required:"true" yaml:"market"`
	Log     logConfig     `required:"true" yaml:"log"`
	Locator locatorConfig `required:"true" yaml:"locator"`
	Keys    keysConfig    `required:"false" yaml:"keys"`
	Hub     *hubConfig    `required:"false" yaml:"hub"`
}

func (y *yamlConfig) ListenAddress() string {
	return y.Node.ListenAddr
}

func (y *yamlConfig) MarketEndpoint() string {
	return y.Market.Endpoint
}

func (y *yamlConfig) LocatorEndpoint() string {
	return y.Locator.Endpoint
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

func (y *yamlConfig) KeyStore() string {
	return y.Keys.Keystore
}

func (y *yamlConfig) PassPhrase() string {
	return y.Keys.Passphrase
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
	ctx     context.Context
	conf    Config
	lis     net.Listener
	srv     *grpc.Server
	privKey *ecdsa.PrivateKey
}

// New creates new Local Node instance
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func New(ctx context.Context, c Config, key *ecdsa.PrivateKey) (*Node, error) {
	lis, err := net.Listen("tcp", c.ListenAddress())
	if err != nil {
		return nil, err
	}

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}
	creds := util.NewTLS(TLSConfig)

	srv := util.MakeGrpcServer(creds)
	// register hub connection if hub addr is set
	if c.HubEndpoint() != "" {
		hub, err := newHubAPI(ctx, c, creds)
		if err != nil {
			return nil, err
		}
		pb.RegisterHubManagementServer(srv, hub)
		log.G(ctx).Info("hub service registered", zap.String("endpt", c.HubEndpoint()))
	}

	market, err := newMarketAPI(ctx, key, c)
	if err != nil {
		return nil, err
	}
	pb.RegisterMarketServer(srv, market)
	log.G(ctx).Info("market service registered", zap.String("endpt", c.MarketEndpoint()))

	deals, err := newDealsAPI(ctx, key)
	if err != nil {
		return nil, err
	}
	pb.RegisterDealManagementServer(srv, deals)
	log.G(ctx).Info("deals service registered")

	tasks, err := newTasksAPI(ctx, key, c, creds)
	if err != nil {
		return nil, err
	}
	pb.RegisterTaskManagementServer(srv, tasks)
	log.G(ctx).Info("tasks service registered")

	return &Node{
		lis:     lis,
		conf:    c,
		ctx:     ctx,
		srv:     srv,
		privKey: key,
	}, nil
}

// Serve binds gRPC services and start it
func (n *Node) Serve() error {
	return n.srv.Serve(n.lis)
}
