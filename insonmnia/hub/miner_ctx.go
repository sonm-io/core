package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/docker/go-connections/nat"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	errSlotNotExists     = errors.New("specified slot does not exist")
	errSlotAlreadyExists = errors.New("specified slot already exists")
	errOrderNotExists    = errors.New("specified order does not exist")
	errForbiddenMiner    = errors.New("miner is forbidden")
)

type OrderID string

func (id OrderID) String() string {
	return string(id)
}

// MinerCtx holds all the data related to a connected Miner
type MinerCtx struct {
	ctx    context.Context
	cancel context.CancelFunc

	// gRPC connection
	grpcConn *grpc.ClientConn
	// gRPC Client
	Client    pb.MinerClient
	statusMap map[string]*pb.TaskStatusReply
	statusMu  sync.Mutex
	// Incoming TCP-connection
	conn net.Conn

	// Miner name received after handshaking.
	uuid string

	// Traffic routing.

	router Router

	// Scheduling.

	mu           sync.Mutex
	capabilities *hardware.Hardware
	usage        *resource.Pool
	usageMapping map[OrderID]resource.Resources
}

func (m *MinerCtx) MarshalJSON() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Marshal(m.usageMapping)
}

func (m *MinerCtx) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Unmarshal(data, &m.usageMapping)
}

func (h *Hub) createMinerCtx(ctx context.Context, conn net.Conn) (*MinerCtx, error) {
	var err error

	if h.creds != nil {
		conn, err = h.tlsHandshake(ctx, conn)
		if err != nil {
			return nil, err
		}
	}

	var (
		m = MinerCtx{
			conn:         conn,
			statusMap:    make(map[string]*pb.TaskStatusReply),
			usageMapping: make(map[OrderID]resource.Resources),
		}
	)
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.grpcConn, err = xgrpc.NewClient(ctx, "miner", nil, grpc.WithDialer(func(_ string, _ time.Duration) (net.Conn, error) {
		return conn, nil
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

func (h *Hub) tlsHandshake(ctx context.Context, conn net.Conn) (net.Conn, error) {
	conn, authInfo, err := h.creds.ClientHandshake(ctx, "", conn)
	if err != nil {
		return nil, err
	}

	switch authInfo := authInfo.(type) {
	case auth.EthAuthInfo:
		if !h.state.ACLHas(authInfo.Wallet.Hex()) {
			return nil, errForbiddenMiner
		}
	default:
		return nil, fmt.Errorf("unsupported AuthInfo %s %T", authInfo.AuthType(), authInfo)
	}

	return conn, nil
}

// ID returns the miner id.
func (m *MinerCtx) ID() string {
	return m.uuid
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

	log.G(m.ctx).Info("received handshake from a Miner")

	capabilities, err := hardware.HardwareFromProto(resp.Capabilities)
	if err != nil {
		log.G(m.ctx).Error("failed to decode capabilities from a Miner", zap.Error(err))
		return err
	}

	log.G(m.ctx).Debug("received Miner's capabilities",
		zap.String("id", resp.Miner),
		zap.Any("capabilities", capabilities),
	)

	m.uuid = resp.Miner
	m.capabilities = capabilities
	m.usage = resource.NewPool(capabilities)

	if m.router, err = h.newRouter(m.uuid, resp.NatType); err != nil {
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
func (h *Hub) newRouter(id string, natType pb.NATType) (Router, error) {
	if h.gateway == nil || natType == pb.NATType_NONE {
		return newDirectRouter(), nil
	}

	if gateway.PlatformSupportIPVS {
		return newIPVSRouter(h.ctx, h.gateway, h.portPool), nil
	}

	return nil, errors.New("miner has firewall configured, but Hub's host OS has no IPVS support")
}

func (m *MinerCtx) deregisterRoute(ID string) error {
	return m.router.Deregister(ID)
}

func (m *MinerCtx) initStatusClient() (statusClient pb.Miner_TasksStatusClient, err error) {
	statusClient, err = m.Client.TasksStatus(m.ctx)
	if err != nil {
		log.G(m.ctx).Error("failed to get status client", zap.Error(err))
		return
	}

	err = statusClient.Send(&pb.MinerStatusMapRequest{})
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

		m.statusMu.Lock()
		m.statusMap = statusReply.Statuses
		m.statusMu.Unlock()
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
			_, err := m.Client.Ping(ctx, &pb.Empty{})
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

// Consume consumes the specified resources from the miner.
func (m *MinerCtx) Consume(Id OrderID, usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.consume(Id, usage)
}

func (m *MinerCtx) consume(id OrderID, usage *resource.Resources) error {
	if m.orderExists(id) {
		return fmt.Errorf("order already exists")
	}
	if err := m.usage.Consume(usage); err != nil {
		return err
	}

	log.G(m.ctx).Debug("consumed resources for a task",
		zap.Stringer("id", id),
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.usage.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	m.usageMapping[OrderID(id)] = *usage

	return nil
}

func (m *MinerCtx) OrderExists(id OrderID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.orderExists(id)
}

func (m *MinerCtx) orderExists(id OrderID) bool {
	_, exists := m.usageMapping[id]
	return exists
}

func (m *MinerCtx) PollConsume(usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.G(m.ctx).Debug("checking for resources",
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.usage.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	return m.usage.PollConsume(usage)
}

// Release returns back resources for the miner.
//
// Should be called when a deal has finished no matter for what reason.
func (m *MinerCtx) Release(id OrderID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.releaseDeal(id)
}

func (m *MinerCtx) releaseDeal(id OrderID) {
	usage, exists := m.usageMapping[id]
	if !exists {
		return
	}

	log.G(m.ctx).Debug("retained resources for a task",
		zap.Stringer("id", id),
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.usage.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	delete(m.usageMapping, id)
	m.usage.Release(&usage)
}

func (m *MinerCtx) OrderUsage(id OrderID) (resource.Resources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orderUsage(id)
}

func (m *MinerCtx) orderUsage(id OrderID) (resource.Resources, error) {
	usage, exists := m.usageMapping[id]
	if !exists {
		return resource.Resources{}, errOrderNotExists
	}

	return usage, nil
}

// Orders returns a list of allocated orders.
// Useful for looking for a proper miner for starting tasks.
func (m *MinerCtx) Orders() []OrderID {
	m.mu.Lock()
	defer m.mu.Unlock()
	orders := []OrderID{}
	for id := range m.usageMapping {
		orders = append(orders, id)
	}
	return orders
}

func (m *MinerCtx) registerRoutes(ID string, routes map[string]*pb.Endpoints) []routeMapping {
	var outRoutes []routeMapping

	for natPort, endpoints := range routes {
		port := nat.Port(natPort)

		vsID := fmt.Sprintf("%s#%s", ID, natPort)
		vs, err := m.router.Register(vsID, port.Proto())
		if err != nil {
			log.G(m.ctx).Warn("failed to register route", zap.Error(err))
			continue
		}

		for _, endpoint := range endpoints.Endpoints {
			outRoute, err := vs.AddReal(vsID, endpoint.Addr, uint16(endpoint.Port))
			if err != nil {
				log.G(m.ctx).Warn("failed to register route", zap.Error(err))
				continue
			}

			outRoutes = append(outRoutes, routeMapping{
				containerPort: port.Port(),
				route:         outRoute,
			})
		}
	}

	return outRoutes
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
	if m.router != nil {
		m.router.Close()
	}
}
