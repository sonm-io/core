package locator

import (
	"fmt"
	"golang.org/x/net/context"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	pb "github.com/sonm-io/core/proto"
	ds "github.com/sonm-io/core/insonmnia/locator/datastruct"
)

type Storage interface {
	ByEthAddr(ethAddr common.Address) (*ds.Node, error)
	Put(n *ds.Node)
}

type Srv struct {
	s Storage
}

func NewServer(s Storage) *Srv {
	return &Srv{s: s}
}

func (s *Srv) Announce(ctx context.Context, req *pb.AnnounceRequest) (*pb.Empty, error) {
	ethAddr, err := ExtractEthAddr(ctx)
	if err != nil {
		return nil, err
	}

	log.G(ctx).Info("Handling Announce request",
		zap.Stringer("eth", ethAddr), zap.Strings("ips", req.IpAddr))

	s.s.Put(&ds.Node{
		EthAddr: ethAddr,
		IpAddr:  req.IpAddr,
	})

	return &pb.Empty{}, nil
}

func (s *Srv) Resolve(ctx context.Context, req *pb.ResolveRequest) (*pb.ResolveReply, error) {
	log.G(ctx).Info("Handling Resolve request", zap.String("eth", req.EthAddr))

	if !common.IsHexAddress(req.EthAddr) {
		return nil, fmt.Errorf("invalid ethaddress %s", req.EthAddr)
	}

	n, err := s.s.ByEthAddr(common.HexToAddress(req.EthAddr))
	if err != nil {
		return nil, err
	}

	return &pb.ResolveReply{IpAddr: n.IpAddr}, nil
}
