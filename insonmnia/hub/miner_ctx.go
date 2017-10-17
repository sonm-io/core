package hub

import (
	"errors"
	"net"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
)

type MinerProperties map[string]string

var (
	errSlotAlreadyExists = errors.New("specified slot already exists")
	errCPUNotEnough      = errors.New("number of CPU cores requested is unable to fit system's capabilities")
	errMemoryNotEnough   = errors.New("number of memory requested is unable to fit system's capabilities")
)

// MinerCtx holds all the data related to a connected Miner
type MinerCtx struct {
	ctx    context.Context
	cancel context.CancelFunc

	// gRPC connection
	grpcConn *grpc.ClientConn
	// gRPC Client
	Client     pb.MinerClient
	status_map map[string]*pb.TaskStatusReply
	status_mu  sync.Mutex
	// Incoming TCP-connection
	conn net.Conn

	// Miner name received after handshaking.
	uuid                string
	capabilities        *hardware.Hardware
	capabilitiesCurrent *resource.Pool
	router              router

	mu    sync.Mutex
	usage map[string]*resource.Resources

	// TODO (3Hren): This is placed here temporarily, because of further scheduling, which currently does not exist.
	minerProperties MinerProperties
	scheduler       Scheduler
}

func (h *Hub) createMinerCtx(ctx context.Context, conn net.Conn) (*MinerCtx, error) {
	var (
		m = MinerCtx{
			conn:            conn,
			status_map:      make(map[string]*pb.TaskStatusReply),
			usage:           make(map[string]*resource.Resources),
			minerProperties: MinerProperties(make(map[string]string)),
		}
		err error
	)
	m.ctx, m.cancel = context.WithCancel(ctx)
	// TODO: secure connection
	// TODO: identify miner via Authorization mechanism
	// TODO: rediscover jobs assigned to that Miner
	dctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	m.grpcConn, err = grpc.DialContext(dctx, "miner", grpc.WithInsecure(),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()), grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDialer(func(_ string, _ time.Duration) (net.Conn, error) {
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

// ID returns the miner id.
func (m *MinerCtx) ID() string {
	return m.uuid
}

func (m *MinerCtx) MinerProperties() MinerProperties {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.minerProperties
}

func (m *MinerCtx) SetMinerProperties(properties MinerProperties) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.minerProperties = properties
}

func (m *MinerCtx) GetSlots() []*structs.Slot {
	return m.scheduler.All()
}

func (m *MinerCtx) AddSlot(slot *structs.Slot) error {
	resources := slot.GetResources()
	if resources.GetCpuCores() > uint64(m.capabilities.LogicalCPUCount()) {
		return errCPUNotEnough
	}
	if resources.GetMemoryInBytes() > m.capabilities.TotalMemory() {
		return errMemoryNotEnough
	}
	return m.scheduler.Add(slot)
}

// ReserveSlot reserves a slot during bid/ask protocol.
func (m *MinerCtx) ReserveSlot(slot *structs.Slot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.reserveSlot(slot)
}

func (m *MinerCtx) reserveSlot(slot *structs.Slot) error {
	return m.scheduler.Reserve(slot)
}

func (m *MinerCtx) HasSlot(slot *structs.Slot) bool {
	return m.scheduler.Exists(slot)
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
	m.capabilitiesCurrent = resource.NewPool(capabilities)

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
func (h *Hub) newRouter(id string, natType pb.NATType) (router, error) {
	if h.gateway == nil || natType == pb.NATType_NONE {
		return newDirectRouter(), nil
	}

	if gateway.PlatformSupportIPVS {
		return newIPVSRouter(h.ctx, id, h.gateway, h.portPool), nil
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
func (m *MinerCtx) Consume(taskID string, usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.consume(taskID, usage)
}

func (m *MinerCtx) consume(taskID string, usage *resource.Resources) error {
	if err := m.capabilitiesCurrent.Consume(usage); err != nil {
		return err
	}

	log.G(m.ctx).Debug("consumed resources for a task",
		zap.String("taskID", taskID),
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.capabilitiesCurrent.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	m.usage[taskID] = usage

	return nil
}

func (m *MinerCtx) PollConsume(usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.capabilitiesCurrent.PollConsume(usage)
}

// Retain retains back resources for the miner.
//
// Should be called when a task has finished no matter for what reason.
func (m *MinerCtx) Retain(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.retain(taskID)
}

func (m *MinerCtx) retain(taskID string) {
	usage, exists := m.usage[taskID]
	if !exists {
		return
	}

	log.G(m.ctx).Debug("retained resources for a task",
		zap.String("taskID", taskID),
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.capabilitiesCurrent.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	delete(m.usage, taskID)
	m.capabilitiesCurrent.Retain(usage)
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
