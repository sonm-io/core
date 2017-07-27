package hub

import (
	"net"
	"time"

	"github.com/hashicorp/yamux"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbminer "github.com/sonm-io/core/proto/miner"
)

// MinerCtx holds all the data related to a connected Miner
type MinerCtx struct {
	ctx    context.Context
	cancel context.CancelFunc

	// gRPC connection
	grpcConn *grpc.ClientConn
	// gRPC Client
	Client     pbminer.MinerClient
	status_map map[string]*pbminer.TaskStatus
	// Incoming TCP-connection
	conn net.Conn

	// TODO: forwarding
	session *yamux.Session
}

func createMinerCtx(ctx context.Context, conn net.Conn) (*MinerCtx, error) {
	var (
		m = MinerCtx{
			conn: conn,
		}
		err error
	)
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.session, err = yamux.Client(conn, nil)
	if err != nil {
		m.Close()
		return nil, err
	}
	yaConn, err := m.session.Open()
	if err != nil {
		m.Close()
		return nil, err
	}
	// TODO: secure connection
	// TODO: identify miner via Authorization mechanism
	// TODO: rediscover jobs assigned to that Miner
	dctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	m.grpcConn, err = grpc.DialContext(dctx, "miner", grpc.WithInsecure(), grpc.WithDialer(func(_ string, _ time.Duration) (net.Conn, error) {
		return yaConn, nil
	}))

	if err != nil {
		log.G(ctx).Error("failed to connect to Miner's grpc server", zap.Error(err))
		m.Close()
		return nil, err
	}

	log.G(ctx).Info("grpc.Dial successfully finished")
	m.Client = pbminer.NewMinerClient(m.grpcConn)
	return &m, nil
}

func (m *MinerCtx) status() error {
	statusClient, err := m.Client.TasksStatus(m.ctx)
	if err != nil {
		return err
	}
	statusClient.Send(&pbminer.TasksStatusRequest{})
	for {
		statusReply, err := statusClient.Recv()
		if err != nil {
			return err
		}
		m.status_map = statusReply.Statuses
	}
	return nil
}

func (m *MinerCtx) ping() error {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			log.G(m.ctx).Info("ping the Miner", zap.Stringer("remote", m.conn.RemoteAddr()))
			// TODO: implement retries
			ctx, cancel := context.WithTimeout(m.ctx, time.Second*5)
			_, err := m.Client.Ping(ctx, &pbminer.PingRequest{})
			cancel()
			if err != nil {
				log.G(ctx).Error("failed to ping miner", zap.Error(err))
				return err
			}
		case <-m.ctx.Done():
			return m.ctx.Err()
		}
	}
}

// Close frees all connections related to a Miner
func (m *MinerCtx) Close() {
	m.cancel()
	if m.grpcConn != nil {
		m.grpcConn.Close()
	}
	if m.conn != nil {
		m.conn.Close()
	}
	if m.session != nil {
		m.session.Close()
	}
}
