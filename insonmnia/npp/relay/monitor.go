package relay

import (
	"net"

	"github.com/hashicorp/memberlist"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type monitor struct {
	cfg MonitorConfig

	server   *grpc.Server
	listener net.Listener

	cluster *memberlist.Memberlist

	log *zap.Logger
}

func newMonitor(cfg MonitorConfig, cluster *memberlist.Memberlist, log *zap.Logger) *monitor {
	server := xgrpc.NewServer(log, xgrpc.DefaultTraceInterceptor())

	m := &monitor{
		cfg:     cfg,
		server:  server,
		cluster: cluster,
	}

	return m
}

func (m *monitor) Cluster(ctx context.Context, request *sonm.Empty) (*sonm.RelayClusterReply, error) {
	membersCluster := m.cluster.Members()
	members := make([]string, 0, len(membersCluster))

	for _, member := range membersCluster {
		members = append(members, member.Address())
	}

	return &sonm.RelayClusterReply{
		Members: members,
	}, nil
}

// TODO: Metrics.

func (m *monitor) Serve() error {
	listener, err := net.Listen("tcp", m.cfg.Endpoint)
	if err != nil {
		return err
	}

	sonm.RegisterRelayServer(m.server, m)
	m.listener = listener

	return m.server.Serve(m.listener)
}

func (m *monitor) Close() error {
	if m.listener != nil {
		m.listener.Close()
	}
	if m.server != nil {
		m.server.Stop()
	}

	return nil
}
