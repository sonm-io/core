// Rendezvous protocol implementation. The other name - Bidirectional Locator.
//
// Rendezvous is a protocol aimed to achieve mutual address resolution for both
// client and server. This is especially useful where there is no guarantee
// that both peers are reachable directly, i.e. behind a NAT for example.
// The protocol allows for servers to publish their private network addresses
// while resolving their real remove address. This information is saved under
// some ID until a connection held and heartbeat frames transmitted.
// When a client wants to perform a connection it goes to this server, informs
// about its own private network addresses and starting from this point a
// rendezvous can be achieved.
// Both client and server are informed about remote peers public and private
// addresses and they may perform several attempts to create an actual p2p
// connection.
//
// By trying to connect to private addresses they can reach each other if and
// only if they're both in the same LAN.
// If a peer doesn't have private address, the connection is almost guaranteed
// to be established directly from the private peer to the public one.
// If both of them are located under different NATs, a TCP punching attempt can
// be performed.
// At last, when there is no hope, a special relay server can be used to
// forward the traffic.
//
// Servers should publish all of possibly reachable endpoints for all protocols
// they support.
// Clients should specify the desired protocol and ID for resolution.
//
// Currently only TCP endpoints exchanging is supported.
// TODO: When resolving it's necessary to track also IP version. For example to be able not to return IPv6 when connecting socket is IPv4.

package rendezvous

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp/nppc"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/sonm-io/core/util/xnet"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type peerCandidate struct {
	Peer
	C chan<- Peer
}

// Meeting describes the meeting room for servers and clients for a specific
// NPP identifier.
type meeting struct {
	mu sync.Mutex
	// Here we have a map where all clients are placed when there are no
	// servers available.
	// We allow multiple clients to be waited for servers.
	clients map[PeerID]*peerCandidate
	// We allow multiple servers can be registered for fault tolerance, but
	// this is unlikely.
	servers map[PeerID]*peerCandidate
	// The timestamp shows when the last server for this ID has been seen.
	// It is updated each time the server de-announces itself and can be
	// safely ignored when the "servers" map has at least one element in it.
	serversLastSeenTime time.Time
}

func newMeeting() *meeting {
	return &meeting{
		servers: map[PeerID]*peerCandidate{},
		clients: map[PeerID]*peerCandidate{},
	}
}

func (m *meeting) AddServer(peer Peer, c chan<- Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.servers[peer.ID] = &peerCandidate{Peer: peer, C: c}
}

func (m *meeting) RemoveServer(id PeerID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.servers, id)

	if len(m.servers) == 0 {
		m.serversLastSeenTime = time.Now()
	}
}

func (m *meeting) AddClient(peer Peer, c chan<- Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[peer.ID] = &peerCandidate{Peer: peer, C: c}
}

func (m *meeting) RemoveClient(id PeerID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.clients, id)
}

func (m *meeting) PopRandomServer() *peerCandidate {
	m.mu.Lock()
	defer m.mu.Unlock()

	candidate := m.popRandomPeerCandidate(m.servers)
	if len(m.servers) == 0 {
		m.serversLastSeenTime = time.Now()
	}

	return candidate
}

func (m *meeting) PopRandomClient() *peerCandidate {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.popRandomPeerCandidate(m.clients)
}

func (m *meeting) popRandomPeerCandidate(candidates map[PeerID]*peerCandidate) *peerCandidate {
	if len(candidates) == 0 {
		return nil
	}
	keys := make([]PeerID, 0, len(candidates))
	for key := range candidates {
		keys = append(keys, key)
	}

	k := keys[rand.Intn(len(keys))]
	v := candidates[k]
	delete(candidates, k)
	return v
}

func (m *meeting) IsServerInactive() bool {
	const expireDuration = 30 * time.Second

	return time.Now().After(m.serversLastSeenTime.Add(expireDuration))
}

// Server represents a rendezvous server.
//
// This server is responsible for tracking servers and clients to make them
// meet each other.
type Server struct {
	cfg         *ServerConfig
	log         *zap.Logger
	server      *grpc.Server
	resolver    *xnet.ExternalPublicIPResolver
	credentials *xgrpc.TransportCredentials
	enableQUIC  bool

	mu sync.Mutex
	rv map[nppc.ResourceID]*meeting
}

// NewServer constructs a new rendezvous server using specified config and
// options.
//
// The server supports TLS by passing transport credentials using
// WithCredentials option.
// Also it is possible to activate logging system by passing a logger using
// WithLogger function as an option.
func NewServer(cfg *ServerConfig, options ...Option) (*Server, error) {
	opts := newOptions()
	for _, option := range options {
		option(opts)
	}

	server := &Server{
		cfg: cfg,
		log: opts.Log,
		server: xgrpc.NewServer(
			opts.Log,
			xgrpc.GRPCServerOptions(
				grpc.KeepaliveParams(keepalive.ServerParameters{
					Time: 30 * time.Second,
				}),
				grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
					MinTime: 30 * time.Second,
				}),
			),
			xgrpc.Credentials(opts.Credentials),
			xgrpc.DefaultTraceInterceptor(),
			xgrpc.RequestLogInterceptor([]string{}),
			xgrpc.VerifyInterceptor(),
		),
		resolver:    xnet.NewExternalPublicIPResolver(""),
		credentials: opts.Credentials,
		enableQUIC:  opts.EnableQUIC,

		rv: map[nppc.ResourceID]*meeting{},
	}

	server.log.Debug("configured authentication settings",
		zap.String("addr", crypto.PubkeyToAddress(cfg.PrivateKey.PublicKey).Hex()),
		zap.Any("credentials", opts.Credentials.Info()),
	)

	sonm.RegisterRendezvousServer(server.server, server)
	server.log.Debug("registered gRPC server")

	return server, nil
}

func (m *Server) Resolve(ctx context.Context, request *sonm.ConnectRequest) (*sonm.RendezvousReply, error) {
	log := logging.WithTrace(ctx, m.log)

	peerInfo, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errNoPeerInfo()
	}

	id := nppc.ResourceID{
		Protocol: request.Protocol,
		Addr:     common.BytesToAddress(request.ID),
	}
	log.Info("resolving remote peer", zap.Stringer("id", id))

	peerHandle := NewPeer(*peerInfo, request.PrivateAddrs)

	c, err := m.addServerWatch(id, peerHandle)
	if err != nil {
		return nil, err
	}
	defer m.removeServerWatch(id, peerHandle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-c:
		log.Info("providing remote server endpoint(s)",
			zap.Stringer("id", id),
			zap.Stringer("public_addr", p.Addr),
			zap.Any("private_addrs", p.privateAddrs),
		)
		return m.newReply(p)
	}
}

func (m *Server) ResolveAll(ctx context.Context, request *sonm.ID) (*sonm.ResolveMetaReply, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var ids []string
	for id, meeting := range m.rv {
		if id.Addr.Hex() == request.Id {
			for peerID := range meeting.servers {
				ids = append(ids, peerID.String())
			}
		}
	}

	if len(ids) == 0 {
		return nil, errPeerNotFound()
	}

	return &sonm.ResolveMetaReply{
		IDs: ids,
	}, nil
}

func (m *Server) Publish(ctx context.Context, request *sonm.PublishRequest) (*sonm.RendezvousReply, error) {
	log := logging.WithTrace(ctx, m.log)

	peerInfo, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	id := nppc.ResourceID{
		Protocol: request.Protocol,
		Addr:     peerInfo.Addr,
	}
	log.Info("publishing remote peer", zap.String("id", id.String()))

	peerHandle := NewPeer(*peerInfo.Peer, request.PrivateAddrs)

	c := m.addClientWatch(id, peerHandle)
	defer m.removeClientWatch(id, peerHandle)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case p := <-c:
		log.Info("providing remote client endpoint(s)",
			zap.Stringer("id", id),
			zap.Stringer("public_addr", p.Addr),
			zap.Any("private_addrs", p.privateAddrs),
		)
		return m.newReply(p)
	}
}

func (m *Server) addServerWatch(id nppc.ResourceID, peer Peer) (<-chan Peer, error) {
	c := make(chan Peer, 2)

	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, ok := m.rv[id]
	if ok {
		// Notify both sides immediately if there is match between candidates.
		if server := meeting.PopRandomServer(); server != nil {
			c <- server.Peer
			server.C <- peer
		} else {
			if meeting.IsServerInactive() {
				return nil, errPeerNotFound()
			}

			meeting.AddClient(peer, c)
		}
	} else {
		return nil, errPeerNotFound()
	}

	return c, nil
}

func (m *Server) addClientWatch(id nppc.ResourceID, peer Peer) <-chan Peer {
	c := make(chan Peer, 2)

	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, ok := m.rv[id]
	if ok {
		if client := meeting.PopRandomClient(); client != nil {
			c <- client.Peer
			client.C <- peer
		} else {
			meeting.AddServer(peer, c)
		}
	} else {
		meeting := newMeeting()
		meeting.AddServer(peer, c)
		m.rv[id] = meeting
	}

	return c
}

func (m *Server) removeClientWatch(id nppc.ResourceID, peer Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	candidates, ok := m.rv[id]
	if ok {
		candidates.RemoveServer(peer.ID)
		m.maybeCleanMeeting(id, candidates)
	}
}

func (m *Server) removeServerWatch(id nppc.ResourceID, peer Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	candidates, ok := m.rv[id]
	if ok {
		candidates.RemoveClient(peer.ID)
		m.maybeCleanMeeting(id, candidates)
	}
}

func (m *Server) maybeCleanMeeting(id nppc.ResourceID, candidates *meeting) {
	if len(candidates.clients) == 0 && len(candidates.servers) == 0 && candidates.IsServerInactive() {
		delete(m.rv, id)
	}
}

func (m *Server) newReply(peer Peer) (*sonm.RendezvousReply, error) {
	addr, err := sonm.NewAddr(peer.Addr)
	if err != nil {
		return nil, err
	}

	// If peer's public IP address suddenly arrived as a private this only
	// means that both Rendezvous and server/client are located within the
	// same VLAN. This should rarely happen, but in these cases we must
	// resolve public IP of ours.
	if addr.IsPrivate() {
		publicIP, err := m.resolver.PublicIP()
		if err != nil {
			return nil, err
		}

		addr.Addr.Addr = publicIP.String()
	}

	return &sonm.RendezvousReply{
		PublicAddr:   addr,
		PrivateAddrs: peer.privateAddrs,
	}, nil
}

func (m *Server) Info(ctx context.Context, request *sonm.Empty) (*sonm.RendezvousState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := make(map[string]*sonm.RendezvousMeeting)

	for id, meeting := range m.rv {
		clients := make(map[string]*sonm.RendezvousReply)
		servers := make(map[string]*sonm.RendezvousReply)

		for clientID, candidate := range meeting.clients {
			reply, err := m.newReply(candidate.Peer)
			if err != nil {
				return nil, err
			}

			clients[clientID.String()] = reply
		}

		for serverID, candidate := range meeting.servers {
			reply, err := m.newReply(candidate.Peer)
			if err != nil {
				return nil, err
			}
			servers[serverID.String()] = reply
		}

		state[id.String()] = &sonm.RendezvousMeeting{
			Servers: servers,
			Clients: clients,
		}
	}

	return &sonm.RendezvousState{
		State: state,
	}, nil
}

// Run starts accepting incoming connections, serving them by blocking the
// caller execution context until either explicitly terminated using Stop
// or some critical error occurred.
//
// Always returns non-nil error.
func (m *Server) Run(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return m.serveGRPCOverTCP(ctx)
	})
	wg.Go(func() error {
		return m.serveGRPCOverQUIC(ctx)
	})
	wg.Go(func() error {
		return m.serveDebug(ctx)
	})

	<-ctx.Done()
	m.Stop()

	return wg.Wait()
}

func (m *Server) serveGRPCOverTCP(ctx context.Context) error {
	log := m.log.Sugar()

	listener, err := net.Listen(m.cfg.Addr.Network(), m.cfg.Addr.String())
	if err != nil {
		return err
	}

	log.Infof("exposing Rendezvous TCP server on %s", listener.Addr().String())
	defer log.Infof("stopped Rendezvous TCP server on %s", listener.Addr().String())

	return m.server.Serve(listener)
}

func (m *Server) serveGRPCOverQUIC(ctx context.Context) error {
	if !m.enableQUIC {
		return nil
	}

	if m.credentials.TLSConfig == nil {
		return fmt.Errorf("QUIC is enabled, but no transport credentials was provided")
	}

	log := m.log.Sugar()

	listener, err := xnet.ListenQUIC("udp", m.cfg.Addr.String(), m.credentials.TLSConfig, xnet.DefaultQUICConfig())
	if err != nil {
		return err
	}

	log.Infof("exposing Rendezvous QUIC server on %s", listener.Addr().String())
	defer log.Infof("stopped Rendezvous QUIC server on %s", listener.Addr().String())

	return m.server.Serve(listener)
}

func (m *Server) serveDebug(ctx context.Context) error {
	if m.cfg.Debug != nil {
		return debug.ServePProf(ctx, *m.cfg.Debug, m.log)
	}

	return nil
}

// Stop stops the server.
//
// It immediately closes all open connections and listeners. Also it cancels
// all active RPCs on the server side and the corresponding pending RPCs on
// the client side will get notified by connection errors.
func (m *Server) Stop() {
	m.log.Info("rendezvous is shutting down")
	m.server.Stop()
}

func errNoPeerInfo() error {
	return errors.New("no peer info provided")
}

func errPeerNotFound() error {
	return status.Errorf(codes.NotFound, "peer not found")
}
