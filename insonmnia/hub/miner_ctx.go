package hub

import (
	"errors"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"sync"

	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/hardware"
	pb "github.com/sonm-io/core/proto"
)

// MinerCtx holds all the data related to a connected Miner
type MinerCtx struct {
	ctx    context.Context
	cancel context.CancelFunc

	// gRPC connection
	grpcConn *grpc.ClientConn
	// gRPC Client
	Client     pb.MinerClient
	status_map map[string]*pb.TaskDetailsReply
	status_mu  sync.Mutex
	// Incoming TCP-connection
	conn net.Conn

	// TODO: forwarding
	session *yamux.Session

	// Miner name received after handshaking.
	uuid         string
	capabilities *hardware.Hardware
	router       router
}

func (h *Hub) createMinerCtx(ctx context.Context, conn net.Conn) (*MinerCtx, error) {
	var (
		m = MinerCtx{
			conn:       conn,
			status_map: make(map[string]*pb.TaskDetailsReply),
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
	m.Client = pb.NewMinerClient(m.grpcConn)

	if err := m.handshake(h); err != nil {
		m.Close()
		return nil, err
	}

	return &m, nil
}

func (m *MinerCtx) handshake(h *Hub) error {
	log.G(m.ctx).Info("sending handshake to a Miner", zap.Stringer("addr", m.conn.RemoteAddr()))
	resp, err := m.Client.Handshake(m.ctx, &pb.MinerHandshakeRequest{})
	if err != nil {
		log.G(m.ctx).Error("failed to receive handshake from a Miner",
			zap.Any("addr", m.conn.RemoteAddr()),
			zap.Error(err),
		)
		return err
	}

	log.G(m.ctx).Info("received handshake from a Miner", zap.Any("resp", resp))

	capabilities, err := hardware.HardwareFromProto(resp.Capabilities)
	if err != nil {
		log.G(m.ctx).Warn("failed to decode capabilities from a Miner", zap.Error(err))
		return err
	}

	m.uuid = resp.Miner
	m.capabilities = capabilities

	if m.router, err = h.newRouter(resp.NatType); err != nil {
		log.G(m.ctx).Warn("failed to create router for a miner",
			zap.String("uuid", m.uuid),
			zap.Error(err),
		)
		// TODO (3Hren): Possible we should disconnect the miner instead. Need investigation.
		m.router = newDirectRouter()
	}

	return nil
}

// NewRouter constructs a new router that will route requests to bypass miner's firewall.
func (h *Hub) newRouter(natType pb.NATType) (router, error) {
	if h.gateway == nil || natType == pb.NATType_NONE {
		return newDirectRouter(), nil
	}

	if gateway.PlatformSupportIPVS {
		return newIPVSRouter(h.ctx, h.gateway, h.portPool), nil
	}

	return nil, errors.New("miner has firewall configured, but Hub's host OS has no IPVS support")
}

func (m *MinerCtx) deregisterRoute(ID string) error {
	return m.router.DeregisterRoute(ID)
}

func (m *MinerCtx) initStatusClient() (statusClient pb.Miner_TasksStatusClient, err error) {
	statusClient, err = m.Client.TasksStatus(m.ctx)
	if err != nil {
		log.G(m.ctx).Error("failed to get status client", zap.Error(err))
		return
	}

	err = statusClient.Send(&pb.TaskDetailsMapRequest{})
	if err != nil {
		log.G(m.ctx).Error("failed to send tasks status request", zap.Error(err))
		return
	}
	return
}

func (m *MinerCtx) pollStatuses() error {
	statusClient, err := m.initStatusClient()
	if err != nil {
		return err
	}

	for {
		statusReply, err := statusClient.Recv()
		if err != nil {
			log.G(m.ctx).Error("failed to receive miner status", zap.Error(err))
			return err
		}

		m.status_mu.Lock()
		m.status_map = statusReply.Statuses
		m.status_mu.Unlock()
	}
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
			_, err := m.Client.Ping(ctx, &pb.PingRequest{})
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
	m.router.Close()
}
