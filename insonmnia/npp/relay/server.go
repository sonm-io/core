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
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type meeting struct {
	conn net.Conn
	tx   chan<- net.Conn
}

type meetingRoom struct {
	mu sync.Mutex
	// Multiple servers can be registered for fault tolerance.
	servers map[nppc.ResourceID]*connPool
	// Tracks the time where servers were active.
	// This is a historical data and is automatically collected when some
	// duration threshold exceeds.
	serverMonitoring map[nppc.ResourceID]time.Time

	log *zap.SugaredLogger
}

func newMeetingRoom(log *zap.Logger) *meetingRoom {
	return &meetingRoom{
		servers:          map[nppc.ResourceID]*connPool{},
		serverMonitoring: map[nppc.ResourceID]time.Time{},
		log:              log.Sugar(),
	}
}

func (m *meetingRoom) ServerActiveTime(addr nppc.ResourceID) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.servers[addr]; ok {
		return time.Now(), nil
	}

	if timestamp, ok := m.serverMonitoring[addr]; ok {
		return timestamp, nil
	}

	return time.Unix(0, 0), fmt.Errorf("server `%s` is neither active nor ever was seen", addr.String())
}

func (m *meetingRoom) PopRandomServer(addr nppc.ResourceID) *meeting {
	m.mu.Lock()
	defer m.mu.Unlock()

	if servers, ok := m.servers[addr]; ok {
		meeting := servers.popRandom()
		if servers.Empty() {
			delete(m.servers, addr)
			m.serverMonitoring[addr] = time.Now()
		}
		return meeting
	}

	return nil
}

func (m *meetingRoom) PopServer(addr nppc.ResourceID, id ConnID) *meeting {
	m.mu.Lock()
	defer m.mu.Unlock()

	if servers, ok := m.servers[addr]; ok {
		meeting := servers.pop(id)
		if servers.Empty() {
			delete(m.servers, addr)
			m.serverMonitoring[addr] = time.Now()
		}
		return meeting
	}
	return nil
}

func (m *meetingRoom) PutServer(addr nppc.ResourceID, id ConnID, conn net.Conn, tx chan<- net.Conn) {
	m.log.Debugf("putting %s server into the meeting map with %s id", addr.String(), id)

	m.mu.Lock()
	defer m.mu.Unlock()

	servers, ok := m.servers[addr]
	if !ok {
		servers = newConnPool()
		m.servers[addr] = servers
		delete(m.serverMonitoring, addr)
	}
	servers.put(id, conn, tx)
}

func (m *meetingRoom) DiscardConnections(addrs []nppc.ResourceID) {
	for _, addr := range addrs {
		m.log.Infof("closing connections associated with %s address", addr.String())

		servers, ok := m.servers[addr]
		if ok {
			for _, server := range servers.candidates {
				server.conn.Close()
			}
		}
		// TODO: We should also properly clean the map after discarding. But check first.
	}
}

func (m *meetingRoom) Info() (map[string]*sonm.RelayMeeting, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	response := map[string]*sonm.RelayMeeting{}
	for id, meeting := range m.servers {
		servers := map[string]*sonm.Addr{}

		for serverID, candidate := range meeting.candidates {
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

type connPool struct {
	candidates map[ConnID]*meeting
}

func newConnPool() *connPool {
	return &connPool{
		candidates: map[ConnID]*meeting{},
	}
}

func (m *connPool) put(id ConnID, conn net.Conn, tx chan<- net.Conn) {
	m.candidates[id] = &meeting{
		conn: conn,
		tx:   tx,
	}
}

func (m *connPool) pop(id ConnID) *meeting {
	v, ok := m.candidates[id]
	if ok {
		delete(m.candidates, id)
		return v
	}

	return nil
}

func (m *connPool) Empty() bool {
	return len(m.candidates) == 0
}

func (m *connPool) popRandom() *meeting {
	if len(m.candidates) == 0 {
		return nil
	}
	var keys []ConnID
	for key := range m.candidates {
		keys = append(keys, key)
	}

	k := keys[rand.Intn(len(keys))]
	v := m.candidates[k]
	delete(m.candidates, k)
	return v
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

	meetingRoom *meetingRoom

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

	meetingRoom := newMeetingRoom(opts.log)
	metrics := newMetrics(meetingRoom)

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
		listener: &BackPressureListener{listener, opts.log},
		cluster:  nil,
		members:  cfg.Cluster.Members,

		meetingRoom: meetingRoom,

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

	// We support both multiple servers and clients.
	switch handshake.PeerType {
	case sonm.PeerType_SERVER:
		tx, rx := mpsc()

		pollCtx, pollCancel := context.WithCancel(ctx)
		defer pollCancel()

		timer := time.NewTimer(m.waitTimeout)
		defer timer.Stop()

		m.continuum.Track(addr) // TODO: We need to stop tracking also.
		m.meetingRoom.PutServer(addr, id, conn, tx)
		defer m.meetingRoom.PopServer(addr, id)

		select {
		case clientConn, ok := <-rx:
			pollCancel()
			if ok {
				if err := sendOk(conn); err != nil {
					return err
				}
				if err := sendOk(clientConn); err != nil {
					return err
				}
				return m.newMeetingHandler(addr).Relay(ctx, conn, clientConn)
			}
		case err := <-m.pollConnectionAsync(pollCtx, conn):
			return err
		case <-timer.C:
			return errTimeout()
		}
	case sonm.PeerType_CLIENT:
		var targetPeer *meeting
		if handshake.HasUUID() {
			targetPeer = m.meetingRoom.PopServer(addr, ConnID(handshake.UUID))
		} else {
			targetPeer = m.meetingRoom.PopRandomServer(addr)
		}

		if targetPeer != nil {
			if err := sendOk(conn); err != nil {
				return err
			}
			if err := sendOk(targetPeer.conn); err != nil {
				return err
			}
			return m.newMeetingHandler(addr).Relay(ctx, targetPeer.conn, conn)
		} else {
			return errNoPeer()
		}
	default:
		return errUnknownType(handshake.PeerType)
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

func (m *server) pollConnectionAsync(ctx context.Context, conn net.Conn) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- m.pollConnection(ctx, conn)
	}()
	return ch
}

func (m *server) pollConnection(ctx context.Context, conn net.Conn) error {
	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			m.log.Debugf("poll connection from %s", conn.RemoteAddr())
			_, err := conn.Read([]byte{})
			if err != nil {
				return err
			}
		}
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
