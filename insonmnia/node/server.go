package node

import (
	"fmt"
	"net"

	"github.com/jinzhu/configor"
	pb "github.com/sonm-io/core/proto"
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

type yamlConfig struct {
	Node   nodeConfig   `required:"true" yaml:"node"`
	Market marketConfig `required:"true" yaml:"market"`
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
	conf Config
	lis  net.Listener
}

// New creates new Local Node instance
func New(c Config) (*Node, error) {
	lis, err := net.Listen("tcp", c.ListenAddress())
	if err != nil {
		return nil, err
	}

	return &Node{
		lis:  lis,
		conf: c,
	}, nil
}

// Serve binds gRPC services and start it
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func (n *Node) Serve() error {
	srv := grpc.NewServer()

	// register hub connection if hub addr is set
	if n.conf.HubEndpoint() != "" {
		hub, err := newHubAPI(n.conf.HubEndpoint())
		if err != nil {
			return err
		}
		pb.RegisterHubManagementServer(srv, hub)
	} else {
		fmt.Printf("hub addr is empty, management interface disabled")
	}

	deals := newDealsAPI()
	pb.RegisterDealManagementServer(srv, deals)

	tasks := newTasksAPI()
	pb.RegisterTaskManagementServer(srv, tasks)

	market := newMarketAPI()
	pb.RegisterMarketServer(srv, market)

	return srv.Serve(n.lis)
}
