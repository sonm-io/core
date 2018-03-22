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
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/netutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type meeting struct {
	conn net.Conn
	tx   chan<- net.Conn
}

type meetingRoom struct {
	mu sync.Mutex
	// We allow multiple clients to be waited for servers.
	clients map[common.Address]*connPool
	// Also we allow the opposite: multiple servers can be registered for
	// fault tolerance.
	servers map[common.Address]*connPool
}

type ConnID string

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

type server struct {
	cfg Config

	port     netutil.Port
	listener net.Listener
	cluster  *memberlist.Memberlist
	members  []string

	mu sync.Mutex
	// We allow multiple clients to be waited for servers.
	clients map[common.Address]*connPool
	// Also we allow the opposite: multiple servers can be registered for
	// fault tolerance.
	servers map[common.Address]*connPool

	continuum *continuum

	handshakeTimeout time.Duration
	waitTimeout      time.Duration
	bufferSize       int

	metrics *metrics

	monitoring *monitor

	log *zap.SugaredLogger
}

// NewServer constructs a new relay server using specified config with
// options.
func NewServer(cfg Config, options ...Option) (*server, error) {
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

	m := &server{
		cfg: cfg,

		port:     port,
		listener: listener,
		cluster:  nil,
		members:  cfg.Cluster.Members,

		clients: map[common.Address]*connPool{},
		servers: map[common.Address]*connPool{},

		continuum: newContinuum(),

		handshakeTimeout: 30 * time.Second,
		waitTimeout:      60 * time.Second,
		bufferSize:       32 * 1024,

		metrics: newMetrics(),

		log: opts.log.Sugar(),
	}

	if err := m.initCluster(cfg.Cluster); err != nil {
		return nil, err
	}

	m.monitoring = newMonitor(cfg.Monitor, m.cluster, opts.log)

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
func (m *server) Serve() error {
	wg := errgroup.Group{}
	wg.Go(m.serveTCP)
	wg.Go(m.serveGRPC)

	return wg.Wait()
}

func (m *server) serveTCP() error {
	m.log.Infof("running Relay server on %s", m.listener.Addr())
	defer m.log.Info("Relay server has been stopped")

	nodes, err := m.cluster.Join(m.members)
	if err != nil {
		return err
	}

	m.log.Infof("joined the cluster of %d nodes", nodes)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		conn, err := m.listener.Accept()
		if err != nil {
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
		m.log.Warnw("failed to process connection", zap.Error(err))

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
		return m.processDiscover(ctx, conn, common.BytesToAddress(handshake.Addr))
	case sonm.PeerType_SERVER, sonm.PeerType_CLIENT:
		return m.processHandshake(ctx, conn, handshake)
	default:
		return errUnknownType(handshake.PeerType)
	}
}

func (m *server) processDiscover(ctx context.Context, conn net.Conn, addr common.Address) error {
	m.log.Debugf("processing discover request %s", conn.RemoteAddr())

	targetAddr, ok := m.continuum.Get(addr)
	if !ok {
		targetAddr = m.cfg.Addr.String()
	}

	m.log.Debugf("redirecting handshake for %s to %s", conn.RemoteAddr(), targetAddr)
	return sendFrame(conn, newDiscoverResponse(targetAddr))
}

func (m *server) processHandshake(ctx context.Context, conn net.Conn, handshake *sonm.HandshakeRequest) error {
	m.log.Debugf("processing handshake request of `%s` type from %s", strings.ToLower(handshake.PeerType.String()), conn.RemoteAddr())

	if err := handshake.Validate(); err != nil {
		return errInvalidHandshake(err)
	}

	timer := time.NewTimer(m.waitTimeout)
	defer timer.Stop()

	tx, rx := mpsc()
	id := ConnID(uuid.New())
	addr := common.BytesToAddress(handshake.Addr)

	// We support both multiple servers and clients.
	switch handshake.PeerType {
	case sonm.PeerType_SERVER:
		// Need to check whether there is a clients awaits us. If so - select
		// a random one and relay.
		// Otherwise put ourselves into a meeting map.

		// TODO: Decompose into "targetPeer := m.meetingRoom.PopRandomClient()"

		var targetPeer *meeting
		clients, ok := m.clients[addr]
		if ok {
			if client := clients.popRandom(); client != nil {
				targetPeer = client
			}
		}

		if targetPeer != nil {
			tx <- targetPeer.conn
		} else {
			m.log.Debug("putting server into the meeting map")
			servers, ok := m.servers[addr]
			if !ok {
				servers = newConnPool()
				m.servers[addr] = servers
			}
			servers.put(id, conn, tx)
		}

		select {
		case clientConn, ok := <-rx:
			if ok {
				if err := sendOk(conn); err != nil {
					return err
				}
				if err := sendOk(clientConn); err != nil {
					return err
				}
				return m.relay(ctx, conn, clientConn)
			}
		case <-timer.C:
			// TODO: m.meetingRoom.PopServer(id)
			if servers, ok := m.servers[addr]; ok {
				servers.pop(id)
			}
			return errTimeout()
		}
	case sonm.PeerType_CLIENT:
		var targetPeer *meeting
		servers, ok := m.servers[addr]
		if ok {
			if handshake.HasUUID() {
				if server := servers.pop(ConnID(handshake.UUID)); server != nil {
					targetPeer = server
				}
			} else {
				if server := servers.popRandom(); server != nil {
					targetPeer = server
				}
			}
		}

		if targetPeer != nil {
			tx <- targetPeer.conn
		} else {
			m.log.Debug("putting client into the meeting map")
			clients, ok := m.clients[addr]
			if !ok {
				clients = newConnPool()
				m.clients[addr] = clients
			}
			clients.put(id, conn, tx)
		}

		select {
		case clientConn, ok := <-rx:
			if ok {
				if err := sendOk(conn); err != nil {
					return err
				}
				if err := sendOk(clientConn); err != nil {
					return err
				}
				return m.relay(ctx, conn, clientConn)
			}
		case <-timer.C:
			if clients, ok := m.clients[addr]; ok {
				clients.pop(id)
			}
			return errTimeout()
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

func (m *server) relay(ctx context.Context, server, client net.Conn) error {
	log := m.log.With(zap.Stringer("server", server.RemoteAddr()), zap.Stringer("client", client.RemoteAddr()))
	log.Info("ready for relaying")
	defer log.Info("finished relaying")

	wg := errgroup.Group{}
	wg.Go(func() error {
		return m.transmitTCP(server, client, log)
	})
	wg.Go(func() error {
		return m.transmitTCP(client, server, log)
	})

	return wg.Wait()
}

func (m *server) transmitTCP(from, to net.Conn, log *zap.SugaredLogger) error {
	// TODO: Accounting.
	buf := make([]byte, m.bufferSize)

	for {
		bytesRead, err := from.Read(buf[:])
		if err != nil {
			return err
		}

		var bytesSent int
		for bytesSent < bytesRead {
			n, err := to.Write(buf[bytesSent:bytesRead])
			if err != nil {
				return err
			}

			bytesSent += n
		}

		log.Debugf("%d bytes transmitted %s -> %s", bytesRead, from.RemoteAddr(), to.RemoteAddr())
	}
}

func (m *server) Close() error {
	m.monitoring.Close()
	return m.listener.Close()
}

func (m *server) NotifyJoin(node *memberlist.Node) {
	m.log.Infof("node `%s` has joined to the cluster from %s", node.Name, node.Address())

	discarded := m.continuum.Add(m.formatEndpoint(node.Addr), 1)
	m.discardConnections(discarded)
}

func (m *server) NotifyLeave(node *memberlist.Node) {
	m.log.Infof("node `%s` has left from the cluster from %s", node.Name, node.Address())

	discarded := m.continuum.Remove(m.formatEndpoint(node.Addr))
	m.discardConnections(discarded)
}

func (m *server) NotifyUpdate(node *memberlist.Node) {
	m.log.Infof("node `%s` has been updated", node.Name)
}

func (m *server) formatEndpoint(ip net.IP) string {
	return fmt.Sprintf("%s:%d", ip.String(), m.port)
}

func (m *server) discardConnections(addrs []common.Address) {
	for _, addr := range addrs {
		m.log.Infof("closing connections associated with %s address", addr.String())

		servers, ok := m.servers[addr]
		if ok {
			for _, server := range servers.candidates {
				server.conn.Close()
			}
		}

		clients, ok := m.clients[addr]
		if ok {
			for _, client := range clients.candidates {
				client.conn.Close()
			}
		}
	}
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
