package locator

import (
	"errors"
	"net"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var errNodeNotFound = errors.New("node with given Eth address cannot be found")

type node struct {
	ethAddr string
	ipAddr  []string
	ts      time.Time
}

type Locator struct {
	mx sync.Mutex

	grpc *grpc.Server
	conf *Config
	db   map[string]*node
	ctx  context.Context
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

	var total = len(l.db)
	var del uint64
	var keep uint64
	for addr, node := range l.db {
		if node.ts.Before(deadline) {
			delete(l.db, addr)
			del++
		} else {
			keep++
		}
	}

	log.G(l.ctx).Debug("expired nodes cleaned",
		zap.Int("total", total), zap.Uint64("keep", keep), zap.Uint64("del", del))
}

func (l *Locator) Announce(ctx context.Context, req *pb.AnnounceRequest) (*pb.Empty, error) {
	log.G(l.ctx).Info("handling Announce request",
		zap.String("eth", req.EthAddr), zap.Strings("ips", req.IpAddr))

	n := &node{
		ethAddr: req.EthAddr,
		ipAddr:  req.IpAddr,
	}

	l.putAnnounce(n)
	return &pb.Empty{}, nil
}

func (l *Locator) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveReply, error) {
	log.G(l.ctx).Info("handling Resolve request", zap.String("eth", req.EthAddr))

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

type Config struct {
	ListenAddr    string
	NodeTTL       time.Duration
	CleanupPeriod time.Duration
}

func DefaultConfig(addr string) *Config {
	return &Config{
		ListenAddr:    addr,
		NodeTTL:       time.Hour,
		CleanupPeriod: time.Minute,
	}
}

func NewLocator(ctx context.Context, conf *Config) *Locator {
	srv := util.MakeGrpcServer(nil)

	l := &Locator{
		db:   make(map[string]*node),
		grpc: srv,
		conf: conf,
		ctx:  ctx,
	}

	go l.cleanExpiredNodes()

	pb.RegisterLocatorServer(srv, l)
	return l
}
