// TODO: Collect metrics.

package npp

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/netutil"
	"go.uber.org/zap"
)

// NATPuncher describes an interface of NAT Punching Protocol.
//
// It should be used to penetrate a NAT while connecting two peers located
// either under the same or different firewalls with network address
// translation enabled.
type NATPuncher interface {
	// Dial dials the given address.
	// Should be used only on client side.
	Dial(addr common.Address) (net.Conn, error)
	// DialContext connects to the address using the provided context.
	// Should be used only on client side.
	DialContext(ctx context.Context, addr common.Address) (net.Conn, error)
	// Accept blocks the current execution context until a new connection
	// arrives.
	//
	// Indented to be used on server side.
	Accept() (net.Conn, error)
	// @antmat said that this method is clearly self-descriptive and much obvious. Wow.
	AcceptContext(ctx context.Context) (net.Conn, error)
	// RemoteAddr returns rendezvous remote address.
	RemoteAddr() net.Addr
	// Close closes the puncher.
	// Any blocked operations will be unblocked and return errors.
	Close() error
}

type natPuncher struct {
	ctx context.Context
	log *zap.Logger

	client          *rendezvousClient
	pending         *lane.Queue
	protocol        string
	listener        net.Listener
	listenerChannel chan connTuple

	maxAttempts int
	timeout     time.Duration
}

func newNATPuncher(ctx context.Context, cfg rendezvous.Config, client *rendezvousClient, proto string, log *zap.Logger) (NATPuncher, error) {
	// It's important here to reuse the Rendezvous client local address for
	// successful NAT penetration in the case of cone NAT.
	listener, err := reuseport.Listen(protocol, client.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	channel := make(chan connTuple, 1)

	m := &natPuncher{
		ctx:             ctx,
		log:             log,
		client:          client,
		pending:         lane.NewQueue(),
		protocol:        proto,
		listenerChannel: channel,
		listener:        listener,

		maxAttempts: cfg.MaxConnectionAttempts,
		timeout:     cfg.Timeout,
	}

	go m.listen()

	return m, nil
}

func (m *natPuncher) listen() error {
	for {
		conn, err := m.listener.Accept()
		m.listenerChannel <- newConnTuple(conn, err)
		switch {
		case err == nil:
		case strings.Contains(err.Error(), "use of closed network connection"):
			return err
		default:
			m.log.Error("failed to listen NPP", zap.Error(err))
			return err
		}
	}
}

func (m *natPuncher) Dial(addr common.Address) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(m.ctx, m.timeout)
	defer cancel()

	return m.DialContext(ctx, addr)
}

func (m *natPuncher) DialContext(ctx context.Context, addr common.Address) (net.Conn, error) {
	addrs, err := m.resolve(ctx, addr)
	if err != nil {
		m.log.Warn("failed to resolve remote peer using rendezvous", zap.Stringer("remote_addr", addr), zap.Error(err))
		return nil, err
	}

	return m.punch(ctx, addrs, clientConnectionWatcher{})
}

func (m *natPuncher) Accept() (net.Conn, error) {
	return m.AcceptContext(m.ctx)
}

func (m *natPuncher) AcceptContext(ctx context.Context) (net.Conn, error) {
	// Check for acceptor listenerChannel, if there is a connection - return immediately.
	select {
	case conn := <-m.listenerChannel:
		m.log.Info("received acceptor peer", zap.Any("conn", conn))
		return conn.unwrap()
	default:
	}

	if conn := m.pending.Dequeue(); conn != nil {
		m.log.Debug("dequeueing pending connection")
		return conn.(net.Conn), nil
	}

	addrs, err := m.publish(ctx)
	if err != nil {
		m.log.Warn("failed to publish itself on the rendezvous", zap.Error(err))
		return nil, newRendezvousError(err)
	}

	m.log.Info("received remote peer endpoints", zap.Any("addrs", *addrs))

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Here the race begins! We're simultaneously trying to connect to ALL
	// provided endpoints with a reasonable timeout. The first winner will
	// be the champion, while others die in agony. Life is cruel.
	conn, err := m.punch(ctx, addrs, &serverConnectionWatcher{Queue: m.pending, Log: m.log})
	if err != nil {
		return nil, newRendezvousError(err)
	}

	return conn, nil
}

func (m *natPuncher) resolve(ctx context.Context, addr common.Address) (*sonm.RendezvousReply, error) {
	privateAddrs, err := m.privateAddrs()
	if err != nil {
		return nil, err
	}

	request := &sonm.ConnectRequest{
		Protocol:     m.protocol,
		PrivateAddrs: []*sonm.Addr{},
		ID:           addr.Bytes(),
	}

	request.PrivateAddrs, err = convertAddrs(privateAddrs)
	if err != nil {
		return nil, err
	}

	return m.client.Resolve(ctx, request)
}

func (m *natPuncher) publish(ctx context.Context) (*sonm.RendezvousReply, error) {
	privateAddrs, err := m.privateAddrs()
	if err != nil {
		return nil, err
	}

	request := &sonm.PublishRequest{
		Protocol:     m.protocol,
		PrivateAddrs: []*sonm.Addr{},
	}

	request.PrivateAddrs, err = convertAddrs(privateAddrs)
	if err != nil {
		return nil, err
	}

	return m.client.Publish(ctx, request)
}

func convertAddrs(addrs []net.Addr) ([]*sonm.Addr, error) {
	var result []*sonm.Addr
	for _, addr := range addrs {
		host, port, err := netutil.SplitHostPort(addr.String())
		if err != nil {
			return nil, err
		}

		result = append(result, &sonm.Addr{
			Protocol: protocol,
			Addr: &sonm.SocketAddr{
				Addr: host.String(),
				Port: uint32(port),
			},
		})
	}

	return result, nil
}

func (m *natPuncher) punch(ctx context.Context, addrs *sonm.RendezvousReply, watcher connectionWatcher) (net.Conn, error) {
	if addrs.Empty() {
		return nil, fmt.Errorf("no addresses resolved")
	}

	channel := make(chan connTuple, 1)
	go m.doPunch(ctx, addrs, channel, watcher)

	select {
	case conn := <-channel:
		return conn.unwrap()
	case conn := <-m.listenerChannel:
		return conn.unwrap()
	}
}

func (m *natPuncher) doPunch(ctx context.Context, addrs *sonm.RendezvousReply, channel chan<- connTuple, watcher connectionWatcher) {

	m.log.Debug("punching", zap.Any("addrs", *addrs))

	pending := make(chan connTuple, 1+len(addrs.PrivateAddrs))
	wg := sync.WaitGroup{}
	wg.Add(len(addrs.PrivateAddrs))

	if addrs.PublicAddr.IsValid() {
		wg.Add(1)

		go func() {
			defer wg.Done()

			conn, err := m.punchAddr(ctx, addrs.PublicAddr)
			m.log.Debug("received NPP NAT connection candidate", zap.Any("remote_addr", *addrs.PublicAddr), zap.Error(err))
			pending <- newConnTuple(conn, err)
		}()
	}

	for _, addr := range addrs.PrivateAddrs {
		go func(addr *sonm.Addr) {
			defer wg.Done()

			conn, err := m.punchAddr(ctx, addr)
			m.log.Debug("received NPP internet connection candidate", zap.Any("remote_addr", *addr), zap.Error(err))
			pending <- newConnTuple(conn, err)
		}(addr)
	}

	go func() {
		wg.Wait()
		close(pending)
	}()

	var peer net.Conn
	var errs = multierror.NewMultiError()
	for conn := range pending {
		m.log.Debug("received NPP connection candidate", zap.Any("remote_addr", conn.RemoteAddr()), zap.Error(conn.err))

		if conn.Error() != nil {
			errs = multierror.AppendUnique(errs, conn.Error())
			continue
		}

		if peer != nil {
			watcher.OnMoreConnections(conn.conn)
		} else {
			peer = conn.conn
			// Do not return here - still need to close possibly successful connections.
			channel <- newConnTuple(peer, nil)
		}
	}

	if peer == nil {
		channel <- newConnTuple(nil, fmt.Errorf("failed to punch the network using NPP: all attempts has failed - %s", errs.Error()))
	}
}

func (m *natPuncher) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	peerAddr, err := addr.IntoTCP()
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	var errs = multierror.NewMultiError()
	for i := 0; i < m.maxAttempts; i++ {
		conn, err = DialContext(ctx, protocol, m.client.LocalAddr().String(), peerAddr.String())
		if err == nil {
			return conn, nil
		}

		errs = multierror.AppendUnique(errs, err)
	}

	return nil, errs.ErrorOrNil()
}

func (m *natPuncher) RemoteAddr() net.Addr {
	return m.client.RemoteAddr()
}

func (m *natPuncher) Close() error {
	var errs []error

	if err := m.listener.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := m.client.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close listener: %+v", errs)
	} else {
		return nil
	}
}

// PrivateAddrs collects and returns private addresses of a network interfaces
// the listening socket bind on.
func (m *natPuncher) privateAddrs() ([]net.Addr, error) {
	return privateAddrs(m.listener.Addr())
}

type connectionWatcher interface {
	OnMoreConnections(conn net.Conn)
}

type serverConnectionWatcher struct {
	Queue *lane.Queue
	Log   *zap.Logger
}

func (m *serverConnectionWatcher) OnMoreConnections(conn net.Conn) {
	m.Log.Debug("enqueueing pending connection")
	m.Queue.Enqueue(conn)
}

type clientConnectionWatcher struct{}

func (clientConnectionWatcher) OnMoreConnections(conn net.Conn) {
	conn.Close()
}
