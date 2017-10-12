package locator

import (
	"errors"
	"net"
	"sync"
	"time"

	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var errNodeNotFound = errors.New("Node with given Eth address cannot be found")

type node struct {
	ethAddr string
	ipAddr  []string
	ts      time.Time
}

type Locator struct {
	grpc *grpc.Server
	conf *LocatorConfig

	mx sync.Mutex
	db map[string]*node
}

func (l *Locator) putAnnounce(n *node) {
	l.mx.Lock()
	defer l.mx.Unlock()

	n.ts = time.Now()
	l.db[n.ethAddr] = n
}

func (l *Locator) getResolve(ethAddr string) (*node, error) {
	l.mx.Lock()
	defer l.mx.Unlock()

	n, ok := l.db[ethAddr]
	if !ok {
		return nil, errNodeNotFound
	}

	return n, nil
}

func (l *Locator) cleanExpiredNodes() {
	t := time.NewTicker(l.conf.CleanupPeriod)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			l.traverseAndClean()
		}
	}
}

func (l *Locator) traverseAndClean() {
	deadline := time.Now().Add(-1 * l.conf.NodeTTL)

	l.mx.Lock()
	defer l.mx.Unlock()

	for addr, node := range l.db {
		if node.ts.Before(deadline) {
			delete(l.db, addr)
		}
	}
}

func (l *Locator) Announce(ctx context.Context, req *pb.AnnounceRequest) (*pb.Empty, error) {
	n := &node{
		ethAddr: req.EthAddr,
		ipAddr:  req.IpAddr,
	}

	l.putAnnounce(n)
	return &pb.Empty{}, nil
}

func (l *Locator) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveReply, error) {
	n, err := l.getResolve(req.EthAddr)
	if err != nil {
		return nil, err
	}

	return &pb.ResolveReply{IpAddr: n.ipAddr}, nil
}

func (l *Locator) Serve() error {
	lis, err := net.Listen("tcp", l.conf.ListenAddr)
	if err != nil {
		return err
	}

	return l.grpc.Serve(lis)
}

type LocatorConfig struct {
	ListenAddr    string
	NodeTTL       time.Duration
	CleanupPeriod time.Duration
}

func DefaultLocatorConfig() *LocatorConfig {
	return &LocatorConfig{
		ListenAddr:    ":9090",
		NodeTTL:       time.Hour,
		CleanupPeriod: time.Minute,
	}
}

func NewLocator(conf *LocatorConfig) *Locator {
	srv := grpc.NewServer(
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
	)

	l := &Locator{
		mx:   sync.Mutex{},
		db:   make(map[string]*node),
		grpc: srv,
		conf: conf,
	}

	go l.cleanExpiredNodes()

	pb.RegisterLocatorServer(srv, l)
	return l
}
