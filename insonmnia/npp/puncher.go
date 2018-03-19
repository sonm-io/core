// TODO: Collect metrics.

package npp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-reuseport"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/netutil"
	"go.uber.org/zap"
)

const (
	maxConnectAttempts = 5
	maxConnectTimeout  = 10 * time.Second
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
	listener        net.Listener
	listenerChannel chan connTuple

	maxAttempts int
	timeout     time.Duration
}

func newNATPuncher(ctx context.Context, client *rendezvousClient) (NATPuncher, error) {
	// It's important here to reuse the Rendezvous client local address for
	// successful NAT penetration in the case of cone NAT.
	listener, err := reuseport.Listen(protocol, client.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	channel := make(chan connTuple, 1)

	m := &natPuncher{
		ctx:             ctx,
		log:             ctxlog.G(ctx),
		client:          client,
		listenerChannel: channel,
		listener:        listener,

		maxAttempts: maxConnectAttempts,
		timeout:     maxConnectTimeout,
	}

	go m.listen()

	return m, nil
}

func (m *natPuncher) listen() error {
	for {
		conn, err := m.listener.Accept()
		m.listenerChannel <- connTuple{conn, err}
		if err != nil {
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
		return nil, err
	}

	return m.punch(ctx, addrs)
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

	addrs, err := m.publish(ctx)
	if err != nil {
		m.log.Error("failed to publish itself on the rendezvous", zap.Error(err))
		return nil, TransportError{err}
	}

	m.log.Info("received remote peer endpoints", zap.Any("addrs", addrs))

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	// Here the race begins! We're simultaneously trying to connect to ALL
	// provided endpoints with a reasonable timeout. The first winner will
	// be the champion, while others die in agony. Life is cruel.
	conn, err := m.punch(ctx, addrs)
	if err != nil {
		return nil, err
	}

	// TODO: At last when there is no hope, use relay server.
	return conn, nil
}

func (m *natPuncher) resolve(ctx context.Context, addr common.Address) (*sonm.RendezvousReply, error) {
	privateAddrs, err := m.privateAddrs()
	if err != nil {
		return nil, err
	}

	request := &sonm.ConnectRequest{
		Protocol:     protocol,
		PrivateAddrs: []*sonm.Addr{},
		ID:           addr.String(),
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

func (m *natPuncher) punch(ctx context.Context, addrs *sonm.RendezvousReply) (net.Conn, error) {
	if addrs.Empty() {
		return nil, fmt.Errorf("no addresses resolved")
	}

	m.log.Debug("punching", zap.Any("addrs", addrs))

	pending := make(chan connTuple, 1+len(addrs.PrivateAddrs))

	if addrs.PublicAddr.IsValid() {
		go func() {
			conn, err := m.punchAddr(ctx, addrs.PublicAddr)
			m.log.Info("using NAT", zap.Any("addr", addrs.PublicAddr), zap.Error(err))
			pending <- newConnTuple(conn, err)
		}()
	}

	for _, addr := range addrs.PrivateAddrs {
		go func(addr *sonm.Addr) {
			conn, err := m.punchAddr(ctx, addr)
			m.log.Info("using private address", zap.Any("addr", addr), zap.Error(err))
			pending <- newConnTuple(conn, err)
		}(addr)
	}

	var errs []error
	var peer net.Conn
	for i := 0; i < 1+len(addrs.PrivateAddrs); i++ {
		conn := <-pending

		if conn.Error() != nil {
			m.log.Info("failed to punch", zap.Error(conn.Error()))
			errs = append(errs, conn.Error())
			continue
		}

		if peer != nil {
			conn.Close()
		} else {
			peer = conn
		}
	}

	if peer != nil {
		return peer, nil
	}

	return nil, fmt.Errorf("failed to punch the network: all attempts has failed - %+v", errs)
}

func (m *natPuncher) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	peerAddr, err := addr.IntoTCP()
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	for i := 0; i < m.maxAttempts; i++ {
		conn, err = DialContext(ctx, protocol, m.client.LocalAddr().String(), peerAddr.String())
		m.log.Debug("failed to punch", zap.Error(err), zap.Int("attempt", i), zap.Stringer("peer_addr", peerAddr))
		if err == nil {
			return conn, nil
		}
	}

	return nil, fmt.Errorf("failed to connect: %s", err)
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
