// This package contains server-side stuff used for relaying connections
// between two peers.
//
// The Relay server is a last resort tool used to penetrate the NAT in case of
// over-paranoid network configurations, such as Symmetrical NAT or manual
// iptables hacks. When NPP gives up it acts as a third-party server which
// transports TCP traffic in the userland.
// Briefly what it allows - is to establish a TCP connection between two hosts
// with private IP addresses.
//
// There are several components in the Relay server, which allows to unite
// several servers into the single cluster, performing client-side
// load-balancing with the help of servers.
//
// First of all, when a peer server wants to publish itself on the Internet it
// connects to the one of known Relay services and sends a DISCOVER request,
// where it provides its own ID - the Ethereum address in our case.
// The Relay, depending on its cluster's state, selects the proper Relay
// service endpoint, where the meeting will be and returns it back.
// Internally all of the discovered Relays have a consistent hash ring, which
// is used to map the ETH address into a point on it to be able to perform load
// balancing.
//
// Consistent hashing is a special kind of hashing such that when a hash table
// is resized, only K/n keys need to be remapped on average, where K is the
// number of keys, and n is the number of slots.
// In contrast, in most traditional hash tables, a change in the number of
// array slots causes nearly all keys to be remapped because the mapping
// between the keys and the slots is defined by a modular operation.
// This allows us to add or remove additional Relay servers depending on their
// load statuses without rebuilding the entire hash ring, which in its case
// requires reconnection for all of the peers in the awaiting-for-other-peer
// state. So the very least part of already connected and awaiting peers will
// be reconnected.
//
// After discovering the proper Relay endpoint a HANDSHAKE message is sent to
// publish the server. Internally an ETH address provided is verified using
// asymmetrical cryptography based on secp256k1 curves.
//
// At the other side the peer client performs almost the same steps, instead of
// its own ETH address it specifies the target ETH address the client wants to
// connect. When at least two peers are discovered the relaying process starts
// by simply copying all the TCP payload without inspection. Thus, an
// authentication between two peers is still required to keep the traffic
// encrypted and avoid MITM attack.
//
// Several Relays can be united in a single cluster by specifying several
// endpoints of other members in the cluster the user want to join. Internally
// a SWIM protocol is used for fast members discovering and convergence. An
// optional message encryption and members authentication can be specified for
// security reasons.
//
// Relay servers obviously require to be hosted on machines with public IP
// address. However additionally an announce endpoint can be specified to host
// Relay servers under the NAT, but with configured PMP or other stuff that
// allows to forward incoming traffic to the private network.

package relay

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/insonmnia/npp/nppc"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/netutil"
	"github.com/sonm-io/core/util/xnet"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type deleter func()

type peerCandidate struct {
	conn net.Conn
	C    chan<- error
}

type meeting struct {
	mu sync.Mutex
	// We allow multiple clients to be waited for servers.
	clients map[ConnID]*peerCandidate
	// Also we allow the opposite: multiple servers can be registered for
	// fault tolerance.
	servers map[ConnID]*peerCandidate
}

func newMeeting() *meeting {
	return &meeting{
		clients: map[ConnID]*peerCandidate{},
		servers: map[ConnID]*peerCandidate{},
	}
}

func (m *meeting) putServer(id ConnID, conn net.Conn, tx chan<- error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.servers[id] = &peerCandidate{
		conn: conn,
		C:    tx,
	}
}

func (m *meeting) putClient(id ConnID, conn net.Conn, tx chan<- error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[id] = &peerCandidate{
		conn: conn,
		C:    tx,
	}
}

func (m *meeting) popRandomServer() *peerCandidate {
	return m.randomPeerCandidate(m.servers)
}

func (m *meeting) popRandomClient() *peerCandidate {
	return m.randomPeerCandidate(m.clients)
}

func (m *meeting) randomPeerCandidate(candidates map[ConnID]*peerCandidate) *peerCandidate {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(candidates) == 0 {
		return nil
	}
	var keys []ConnID
	for key := range candidates {
		keys = append(keys, key)
	}

	k := keys[rand.Intn(len(keys))]
	v := candidates[k]
	delete(candidates, k)
	return v
}

type meetingHandlerFactory func(addr nppc.ResourceID) *meetingHandler

type meetingHall struct {
	mu                sync.Mutex
	newMeetingHandler meetingHandlerFactory
	// Meeting room for each common address.
	meetingRoom map[nppc.ResourceID]*meeting
	// Map of server connections being active last time. The time is updated
	// each time either a connection arrives or being taken for serving.
	backlog map[nppc.ResourceID]time.Time

	log *zap.SugaredLogger
}

func newMeetingHall(newMeetingHandler meetingHandlerFactory, log *zap.Logger) *meetingHall {
	return &meetingHall{
		newMeetingHandler: newMeetingHandler,
		meetingRoom:       map[nppc.ResourceID]*meeting{},
		backlog:           map[nppc.ResourceID]time.Time{},
		log:               log.Sugar(),
	}
}

func (m *meetingHall) addServerWatch(ctx context.Context, id nppc.ResourceID, connID ConnID, conn net.Conn) <-chan error {
	c := make(chan error, 2)

	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, ok := m.meetingRoom[id]
	if ok {
		// Notify both sides immediately if there is match between candidates.
		if server := meeting.popRandomServer(); server != nil {
			m.log.Infow("providing remote server", zap.Stringer("remoteAddr", server.conn.RemoteAddr()))

			c <- nil
			server.C <- nil

			go func() {
				err := m.executeMeeting(ctx, id, server.conn, conn)

				// Notify both watchers.
				c <- err
				server.C <- err
			}()
		} else {
			meeting.putClient(connID, conn, c)
		}
	} else {
		meeting := newMeeting()
		meeting.putClient(connID, conn, c)
		m.meetingRoom[id] = meeting
	}

	return c
}

func (m *meetingHall) addClientWatch(ctx context.Context, id nppc.ResourceID, connID ConnID, conn net.Conn) <-chan error {
	c := make(chan error, 2)

	m.mu.Lock()
	defer m.mu.Unlock()

	meeting, ok := m.meetingRoom[id]
	if ok {
		if client := meeting.popRandomClient(); client != nil {
			m.log.Infow("providing remote client", zap.Stringer("remoteAddr", client.conn.RemoteAddr()))

			client.C <- nil
			c <- nil

			go func() {
				err := m.executeMeeting(ctx, id, conn, client.conn)

				// Notify both watchers.
				client.C <- err
				c <- err
			}()
		} else {
			meeting.putServer(connID, conn, c)
		}
	} else {
		meeting := newMeeting()
		meeting.putServer(connID, conn, c)
		m.meetingRoom[id] = meeting
	}

	return c
}

func (m *meetingHall) executeMeeting(ctx context.Context, id nppc.ResourceID, server, client net.Conn) error {
	if err := m.finishHandshake(server, client); err != nil {
		return err
	}

	return m.newMeetingHandler(id).Relay(ctx, server, client)
}

func (m *meetingHall) finishHandshake(server, client net.Conn) error {
	if err := sendOk(server); err != nil {
		return err
	}
	if err := sendOk(client); err != nil {
		return err
	}

	return nil
}

func (m *meetingHall) removeClientWatch(id nppc.ResourceID, connID ConnID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	candidates, ok := m.meetingRoom[id]
	if ok {
		delete(candidates.servers, connID)
	}

	m.maybeCleanMeeting(id, candidates)
}

func (m *meetingHall) removeServerWatch(id nppc.ResourceID, connID ConnID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	candidates, ok := m.meetingRoom[id]
	if ok {
		delete(candidates.clients, connID)
	}

	m.maybeCleanMeeting(id, candidates)
}

func (m *meetingHall) maybeCleanMeeting(id nppc.ResourceID, candidates *meeting) {
	if candidates == nil {
		return
	}

	if len(candidates.clients) == 0 && len(candidates.servers) == 0 {
		delete(m.meetingRoom, id)
	}
}

func (m *meetingHall) DiscardConnections(addrs []nppc.ResourceID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, addr := range addrs {
		m.log.Infof("closing connections associated with %s address", addr.String())

		meetingRoom, ok := m.meetingRoom[addr]
		if ok {
			for _, server := range meetingRoom.servers {
				server.conn.Close()
			}
			for _, client := range meetingRoom.clients {
				client.conn.Close()
			}
		}
	}
}

func (m *meetingHall) Info() (map[string]*sonm.RelayMeeting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	response := map[string]*sonm.RelayMeeting{}
	for id, meeting := range m.meetingRoom {
		servers := map[string]*sonm.Addr{}

		for serverID, candidate := range meeting.servers {
			addr, err := sonm.NewAddr(candidate.conn.RemoteAddr())
			if err != nil {
				return nil, err
			}

			servers[serverID.String()] = addr
		}

		response[id.String()] = &sonm.RelayMeeting{Servers: servers}
	}

	return response, nil
}

type ConnID string

func (m ConnID) String() string {
	return string(m)
}

type meetingHandler struct {
	bufferSize int
	metrics    *netMetrics
	log        *zap.SugaredLogger
}

func (m *meetingHandler) Relay(ctx context.Context, server, client net.Conn) error {
	log := m.log.With(zap.Stringer("server", server.RemoteAddr()), zap.Stringer("client", client.RemoteAddr()))
	log.Info("ready for relaying")
	defer log.Info("finished relaying")

	wg := errgroup.Group{}
	wg.Go(func() error {
		return m.transmitTCP(server, client, m.metrics.TxBytes, log)
	})
	wg.Go(func() error {
		return m.transmitTCP(client, server, m.metrics.RxBytes, log)
	})

	return wg.Wait()
}

func (m *meetingHandler) transmitTCP(from, to net.Conn, metrics *atomic.Uint64, log *zap.SugaredLogger) error {
	defer from.(*net.TCPConn).CloseRead()
	defer to.(*net.TCPConn).CloseWrite()

	buf := make([]byte, m.bufferSize)

	for {
		bytesRead, errRead := from.Read(buf[:])
		if bytesRead > 0 {
			var bytesSent int
			for bytesSent < bytesRead {
				n, err := to.Write(buf[bytesSent:bytesRead])
				if err != nil {
					return err
				}

				bytesSent += n
				metrics.Add(uint64(n))
			}

			log.Debugf("%d bytes transmitted %s -> %s", bytesRead, from.RemoteAddr(), to.RemoteAddr())
		}

		if errRead != nil {
			return errRead
		}
	}
}

type server struct {
	cfg ServerConfig

	port     netutil.Port
	listener net.Listener
	cluster  *memberlist.Memberlist
	members  []string

	mu          sync.Mutex
	meetingRoom *meetingHall

	continuum *continuum

	handshakeTimeout time.Duration
	waitTimeout      time.Duration

	metrics           *metrics
	newMeetingHandler func(addr nppc.ResourceID) *meetingHandler

	monitoring *monitor

	log *zap.SugaredLogger
}

// NewServer constructs a new relay server using specified config with
// options.
func NewServer(cfg ServerConfig, options ...Option) (*server, error) {
	opts := newOptions()

	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	opts.log.Sugar().Debugw("configuring Relay server", zap.Any("cfg", cfg))

	port, err := netutil.ExtractPort(cfg.Addr.String())
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen(cfg.Addr.Network(), cfg.Addr.String())
	if err != nil {
		return nil, err
	}

	metrics := newMetrics()

	newMeetingHandler := func(addr nppc.ResourceID) *meetingHandler {
		return &meetingHandler{
			bufferSize: opts.bufferSize,

			metrics: metrics.NetMetrics(addr),
			log:     opts.log.Sugar(),
		}
	}

	m := &server{
		cfg: cfg,

		port:     port,
		listener: &xnet.BackPressureListener{Listener: listener, Log: opts.log},
		cluster:  nil,
		members:  cfg.Cluster.Members,

		meetingRoom: newMeetingHall(newMeetingHandler, opts.log),

		continuum: newContinuum(),

		handshakeTimeout: 30 * time.Second,
		waitTimeout:      24 * time.Hour,

		metrics:           metrics,
		newMeetingHandler: newMeetingHandler,

		log: opts.log.Sugar(),
	}

	if err := m.initCluster(cfg.Cluster); err != nil {
		return nil, err
	}

	m.monitoring, err = newMonitor(cfg.Monitor, m.cluster, m.meetingRoom, metrics, opts.log)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *server) initCluster(cfg ClusterConfig) error {
	key, err := hex.DecodeString(cfg.SecretKey)
	if err != nil {
		return err
	}

	keyring, err := memberlist.NewKeyring([][]byte{}, key)
	if err != nil {
		return err
	}

	addr, port, err := netutil.SplitHostPort(cfg.Endpoint)
	if err != nil {
		return err
	}

	config := memberlist.DefaultWANConfig()
	config.Name = cfg.Name
	config.BindAddr = addr.String()
	config.BindPort = int(port)

	if len(cfg.Announce) > 0 {
		announceAddr, announcePort, err := netutil.SplitHostPort(cfg.Announce)
		if err != nil {
			return err
		}

		config.AdvertiseAddr = announceAddr.String()
		config.AdvertisePort = int(announcePort)
	}
	config.Events = m
	config.Keyring = keyring
	config.LogOutput = newLogAdapter(m.log.Desugar())
	config.ProbeInterval = time.Second

	m.cluster, err = memberlist.Create(config)
	if err != nil {
		return err
	}

	return nil
}

// Serve starts the relay TCP server.
func (m *server) Serve(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return m.serveTCP(ctx)
	})
	wg.Go(func() error {
		// GRPC API doesn't allow to forward the context. Hence the server
		// must be stopped explicitly.
		return m.serveGRPC()
	})
	if m.cfg.Debug != nil {
		wg.Go(func() error {
			return debug.ServePProf(ctx, *m.cfg.Debug, m.log.Desugar())
		})
	}

	<-ctx.Done()
	m.Close()

	return wg.Wait()
}

func (m *server) serveTCP(ctx context.Context) error {
	m.log.Infof("running TCP Relay server on %s", m.listener.Addr())
	defer m.log.Info("TCP Relay server has been stopped")

	nodes, err := m.cluster.Join(m.members)
	if err != nil {
		return err
	}

	m.log.Infof("joined the cluster of %d nodes", nodes)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		conn, err := m.listener.Accept()
		if err != nil {
			m.log.Warnf("failed to accept connection: %v", err)
			return err
		}

		m.log.Debugf("accepted connection from %s", conn.RemoteAddr())
		go m.processConnection(ctx, conn)
	}
}

func (m *server) serveGRPC() error {
	return m.monitoring.Serve()
}

func (m *server) processConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer m.log.Debugf("done processing connection %s", conn.RemoteAddr().String())

	m.metrics.ConnCurrent.Inc()
	defer m.metrics.ConnCurrent.Dec()

	if err := m.processConnectionBlocking(ctx, conn); err != nil {
		if err != io.EOF {
			m.log.Warnw("failed to process connection", zap.Error(err))
		}

		switch e := err.(type) {
		case *protocolError:
			if err := sendError(conn, e.code, e.description); err != nil {
				m.log.Warnw("failed to send error reply", zap.Error(err))
			}
		case error:
			// Do nothing.
		}
	}
}

func (m *server) processConnectionBlocking(ctx context.Context, conn net.Conn) error {
	// Then we need to read a handshake message. Until this, we don't know
	// whether the connection is a server or a client.
	ctx, cancel := context.WithTimeout(ctx, m.handshakeTimeout)
	defer cancel()

	handshake, err := m.readHandshake(ctx, conn)
	if err != nil {
		return err
	}

	switch handshake.PeerType {
	case sonm.PeerType_DISCOVER:
		addr := nppc.ResourceID{
			Protocol: handshake.Protocol,
			Addr:     common.BytesToAddress(handshake.Addr),
		}
		return m.processDiscover(ctx, conn, addr)
	case sonm.PeerType_SERVER, sonm.PeerType_CLIENT:
		return m.processHandshake(ctx, conn, handshake)
	default:
		return errUnknownType(handshake.PeerType)
	}
}

func (m *server) processDiscover(ctx context.Context, conn net.Conn, addr nppc.ResourceID) error {
	m.log.Debugf("processing discover request %s", conn.RemoteAddr())

	targetNode, err := m.continuum.GetNode(addr)
	if err != nil {
		return err
	}

	m.log.Debugf("redirecting handshake for %s to %s", conn.RemoteAddr(), targetNode.String())
	return sendFrame(conn, newDiscoverResponse(targetNode.Addr))
}

func (m *server) processHandshake(ctx context.Context, conn net.Conn, handshake *sonm.HandshakeRequest) error {
	m.log.Debugf("processing handshake request of `%s` type from %s", strings.ToLower(handshake.PeerType.String()), conn.RemoteAddr())

	if err := handshake.Validate(); err != nil {
		return errInvalidHandshake(err)
	}

	id := ConnID(uuid.New())
	addr := nppc.ResourceID{
		Protocol: handshake.Protocol,
		Addr:     common.BytesToAddress(handshake.Addr),
	}

	targetNode, err := m.continuum.GetNode(addr)
	if err != nil {
		return err
	}

	// Peer might have got a no longer valid node address while discovery.
	if targetNode.Name != m.cfg.Cluster.Name {
		return errWrongNode()
	}

	m.log.Infow("publishing remote peer", zap.String("id", id.String()))

	var rx <-chan error

	// We support both multiple servers and clients.
	switch handshake.PeerType {
	case sonm.PeerType_SERVER:
		timer := time.NewTimer(m.waitTimeout)
		defer timer.Stop()

		m.continuum.Track(addr) // TODO: Also undo tracking.

		rx = m.meetingRoom.addClientWatch(ctx, addr, id, conn)
		defer m.meetingRoom.removeClientWatch(addr, id)

		select {
		case <-timer.C:
			return errTimeout()
		case err := <-rx:
			if err != nil {
				return err
			}
		}
	case sonm.PeerType_CLIENT:
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()

		rx = m.meetingRoom.addServerWatch(ctx, addr, id, conn)
		defer m.meetingRoom.removeServerWatch(addr, id)

		select {
		case <-timer.C:
			return errTimeout()
		case err := <-rx:
			if err != nil {
				return err
			}
		}
	default:
		return errUnknownType(handshake.PeerType)
	}

	if err := <-rx; err != nil {
		return err
	}

	return nil
}

func (m *server) readHandshake(ctx context.Context, conn net.Conn) (*sonm.HandshakeRequest, error) {
	channel := make(chan interface{})

	go func() {
		handshake := &sonm.HandshakeRequest{}
		err := recvFrame(conn, handshake)
		if err == nil {
			if handshake.Protocol == "" {
				handshake.Protocol = sonm.DefaultNPPProtocol
			}

			channel <- handshake
		} else {
			channel <- err
		}
	}()

	select {
	case v := <-channel:
		switch v.(type) {
		case *sonm.HandshakeRequest:
			return v.(*sonm.HandshakeRequest), nil
		case error:
			return nil, v.(error)
		default:
			return nil, fmt.Errorf("invalid handshake message: %T", v)
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *server) Close() error {
	m.monitoring.Close()
	return m.listener.Close()
}

func (m *server) NotifyJoin(node *memberlist.Node) {
	m.log.Infof("node `%s` has joined to the cluster from %s", node.Name, node.Address())

	continuumNode, err := newNode(node.Name, m.formatEndpoint(node.Addr))
	if err != nil {
		m.log.Warnf("received malformed node join notification: %v", err)
		return
	}

	discarded := m.continuum.Add(continuumNode.String(), 1)
	m.meetingRoom.DiscardConnections(discarded)
}

func (m *server) NotifyLeave(node *memberlist.Node) {
	m.log.Infof("node `%s` has left from the cluster from %s", node.Name, node.Address())

	continuumNode, err := newNode(node.Name, m.formatEndpoint(node.Addr))
	if err != nil {
		m.log.Warnf("received malformed node leave notification: %v", err)
		return
	}

	discarded := m.continuum.Remove(continuumNode.String())
	m.meetingRoom.DiscardConnections(discarded)
}

func (m *server) NotifyUpdate(node *memberlist.Node) {
	m.log.Infof("node `%s` has been updated", node.Name)
}

func (m *server) formatEndpoint(ip net.IP) string {
	addr := net.TCPAddr{
		IP:   ip,
		Port: int(m.port),
	}

	return addr.String()
}

func mpsc() (chan<- net.Conn, <-chan net.Conn) {
	txrx := make(chan net.Conn, 1)
	return txrx, txrx
}

func sendOk(conn net.Conn) error {
	return sendError(conn, 0, "")
}

func sendError(conn net.Conn, code int32, description string) error {
	response := &sonm.HandshakeResponse{
		Error:       code,
		Description: description,
	}
	return sendFrame(conn, response)
}
