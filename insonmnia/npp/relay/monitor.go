package relay

import (
	"net"

	"github.com/hashicorp/memberlist"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type monitor struct {
	cfg MonitorConfig

	certificate util.HitlessCertRotator
	server      *grpc.Server
	listener    net.Listener

	cluster *memberlist.Memberlist

	metrics *metrics
	log     *zap.Logger
}

func newMonitor(cfg MonitorConfig, cluster *memberlist.Memberlist, metrics *metrics, log *zap.Logger) (*monitor, error) {
	certificate, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	credentials := util.NewTLS(TLSConfig)

	server := xgrpc.NewServer(log,
		xgrpc.Credentials(credentials),
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.VerifyInterceptor(),
	)

	m := &monitor{
		cfg:         cfg,
		certificate: certificate,
		server:      server,
		cluster:     cluster,
		metrics:     metrics,
		log:         log,
	}

	return m, nil
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

func (m *monitor) Metrics(ctx context.Context, request *sonm.Empty) (*sonm.RelayMetrics, error) {
	return m.metrics.Dump(), nil
}

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
	m.server.Stop()
	m.certificate.Close()

	return nil
}
