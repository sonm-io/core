package npp

import (
	"context"
	"sync"
	"time"

	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/uber-go/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GRPCDialer is a cached gRPC client connection handler.
type GRPCDialer struct {
	dialer *Dialer
	mu     sync.Mutex
	cache  *connectionCache
}

func NewGRPCDialer(dialer *Dialer) *GRPCDialer {
	return &GRPCDialer{
		dialer: dialer,
		cache:  newConnectionCache(),
	}
}

func (m *GRPCDialer) RunCG(ctx context.Context) {
	m.cache.RunGC(ctx)
}

func (m *GRPCDialer) Connect(ctx context.Context, addr auth.Addr, credentials credentials.TransportCredentials) (*grpc.ClientConn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cachedGRPCConn, ok := m.cache.Get(addr)
	if ok {
		// TODO: check source. if relay - try to reconnect.
		return cachedGRPCConn, nil
	}

	ethAddr, err := addr.ETH()
	if err != nil {
		return nil, err
	}

	conn, err := m.dialer.DialContextNPP(ctx, addr)
	if err != nil {
		return nil, err
	}

	cachedConn := newCachedConn(conn, m.cache)

	grpcConn, err := xgrpc.NewClient(ctx, "-", auth.NewWalletAuthenticator(credentials, ethAddr), xgrpc.WithConn(cachedConn))
	if err != nil {
		return nil, err
	}

	m.cache.Insert(addr, grpcConn, cachedConn)
	return grpcConn, nil
}

type connectionTuple struct {
	NPP  *cachedConn
	GRPC *grpc.ClientConn
}

type connectionCache struct {
	mu      sync.Mutex
	entries map[string]*connectionTuple
}

func newConnectionCache() *connectionCache {
	return &connectionCache{
		entries: map[string]*connectionTuple{},
	}
}

func (m *connectionCache) RunGC(ctx context.Context) {
	go m.runGC(ctx)
}

func (m *connectionCache) runGC(ctx context.Context) {
	timer := time.NewTicker(1 * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			m.collect()
		}
	}
}

func (m *connectionCache) collect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for addr, conn := range m.entries {
		if time.Since(conn.NPP.activeTime) >= 15*time.Minute {
			conn.GRPC.Close()
			delete(m.entries, addr)
		}
	}
}

func (m *connectionCache) Get(addr auth.Addr) (*grpc.ClientConn, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.entries[addr.String()]
	if conn != nil {
		if conn.NPP.InUse.Inc() == 1 {
			// We're too late.
			return nil, false
		}

		return conn.GRPC, ok
	}

	return nil, ok
}

func (m *connectionCache) Insert(addr auth.Addr, conn *grpc.ClientConn, cachedConn *cachedConn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries[addr.String()] = &connectionTuple{NPP: cachedConn, GRPC: conn}
}

func (m *connectionCache) Remove(addr auth.Addr) *connectionTuple {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := m.entries[addr.String()]
	delete(m.entries, addr.String())
	return conn
}

type cachedConn struct {
	*NPPConn

	InUse *atomic.Uint32

	cache      *connectionCache
	createTime time.Time
	activeTime time.Time
}

func newCachedConn(conn *NPPConn, cache *connectionCache) *cachedConn {
	return &cachedConn{
		NPPConn:    conn,
		InUse:      atomic.NewUint32(1),
		cache:      cache,
		createTime: time.Now(),
		activeTime: time.Now(),
	}
}

func (m *cachedConn) Read(b []byte) (int, error) {
	m.activeTime = time.Now()

	n, err := m.NPPConn.Read(b)
	if err != nil {
		m.forceClose()
	}

	return n, err
}

func (m *cachedConn) Write(b []byte) (int, error) {
	m.activeTime = time.Now()

	n, err := m.NPPConn.Write(b)
	if err != nil {
		m.forceClose()
	}

	return n, err
}

func (m *cachedConn) Close() error {
	if m.InUse.CAS(1, 0) {
		m.forceClose()
		return m.NPPConn.Close()
	}

	return nil
}

func (m *cachedConn) forceClose() {
	if conn := m.cache.Remove(m.Addr); conn != nil {
		conn.GRPC.Close()
	}
}
