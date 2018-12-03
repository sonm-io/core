package npp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lucas-clemente/quic-go"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xnet"
	"go.uber.org/zap"
	"gopkg.in/oleiade/lane.v1"
)

const (
	transportProtocol = "quic"
)

type quicPuncher struct {
	rendezvousClient *rendezvousClient
	tlsConfig        *tls.Config

	listenerChannel    chan connTuple
	pendingConnections *lane.Queue
	protocol           string

	timeout time.Duration
	log     *zap.SugaredLogger
}

func newQUICPuncher(rendezvousClient *rendezvousClient, tlsConfig *tls.Config, protocol string, log *zap.Logger) (*quicPuncher, error) {
	// TODO: Le soutien.
	if protocol == sonm.DefaultNPPProtocol {
		protocol = "grpc"
	}

	m := &quicPuncher{
		rendezvousClient: rendezvousClient,
		tlsConfig:        tlsConfig,

		listenerChannel:    make(chan connTuple, 1),
		pendingConnections: lane.NewQueue(),
		protocol:           strings.Join([]string{transportProtocol, protocol}, "+"),

		timeout: 30 * time.Second,
		log:     log.With(zap.String("type", "QUIC")).Sugar(),
	}

	go func() {
		if err := m.listen(); err != nil {
			m.log.Warn("QUIC listener is closed", zap.Error(err))
		}
	}()

	return m, nil
}

func (m *quicPuncher) listen() error {
	conn := m.rendezvousClient.UDPConn

	m.log.Debugf("exposing QUIC listener on %s", conn.LocalAddr().String())
	defer m.log.Debugf("finished QUIC listening on %s", conn.LocalAddr().String())

	listener, err := quic.Listen(conn, m.tlsConfig, xnet.DefaultQUICConfig())
	if err != nil {
		return err
	}

	wrappedListener := xnet.QUICListener{Listener: listener}

	for {
		conn, err := wrappedListener.Accept()
		m.listenerChannel <- newConnTuple(conn, err)
		switch {
		case err == nil:
			//case strings.Contains(err.Error(), "use of closed network connection"):
			//	return err
		default:
			m.log.Errorw("failed to listen QUIC NPP", zap.Error(err))
			return err
		}
	}
}

func (m *quicPuncher) Accept() (net.Conn, error) {
	return m.AcceptContext(context.Background())
}

func (m *quicPuncher) AcceptContext(ctx context.Context) (net.Conn, error) {
	// Check for pending connections in the acceptor from the channel, if
	// there is a connection - return immediately.
	select {
	case conn := <-m.listenerChannel:
		if err := conn.Error(); err == nil {
			m.log.Infof("received acceptor peer from %s", conn.RemoteAddr())
		}

		return conn.unwrap()
	default:
	}

	if conn := m.pendingConnections.Dequeue(); conn != nil {
		m.log.Debugf("dequeueing pending connection")
		return conn.(net.Conn), nil
	}

	addrs, err := m.publish(ctx)
	if err != nil {
		m.log.Warnw("failed to publish itself on the rendezvous", zap.Error(err))
		return nil, newRendezvousError(err)
	}

	m.log.Infof("received remote peer endpoints: %s", *addrs)

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	conn, err := m.punch(ctx, addrs, &serverConnectionWatcher{Queue: m.pendingConnections, Log: m.log.Desugar()}, true)
	if err != nil {
		return nil, newRendezvousError(err)
	}

	return conn, nil
}

func (m *quicPuncher) publish(ctx context.Context) (*sonm.RendezvousReply, error) {
	request := &sonm.PublishRequest{
		Protocol: m.protocol,
	}

	return m.rendezvousClient.Publish(ctx, request)
}

func (m *quicPuncher) punch(ctx context.Context, addrs *sonm.RendezvousReply, watcher connectionWatcher, isServer bool) (net.Conn, error) {
	if addrs.Empty() {
		return nil, fmt.Errorf("no addresses resolved")
	}

	channel := make(chan connTuple, 1)
	go m.doPunch(ctx, addrs, channel, watcher)

	// todo: out of band conns.
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case conn := <-channel:
			if !isServer {
				return conn.unwrap()
			}
		case conn := <-m.listenerChannel:
			if isServer {
				return conn.unwrap()
			}
			// todo: what to do if not?
		}
	}
}

func (m *quicPuncher) doPunch(ctx context.Context, addrs *sonm.RendezvousReply, channel chan<- connTuple, watcher connectionWatcher) {
	m.log.Debugf("punching %s", *addrs)

	conn, err := m.punchAddr(ctx, addrs.PublicAddr)
	if err != nil {
		channel <- newConnTuple(nil, err)
		return
	}

	m.log.Debugf("received NPP NAT connection candidate: %s", *addrs.PublicAddr)
	channel <- newConnTuple(conn, nil)
}

func (m *quicPuncher) punchAddr(ctx context.Context, addr *sonm.Addr) (net.Conn, error) {
	peerAddr, err := addr.IntoUDP()
	if err != nil {
		return nil, err
	}

	udpConn := m.rendezvousClient.UDPConn

	cfg := xnet.DefaultQUICConfig()

	session, err := quic.DialContext(ctx, udpConn, peerAddr, peerAddr.String(), m.tlsConfig, cfg)
	if err != nil {
		return nil, err
	}

	return xnet.NewQUICConn(session)
}

func (m *quicPuncher) Dial(addr common.Address) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	return m.DialContext(ctx, addr)
}

func (m *quicPuncher) DialContext(ctx context.Context, addr common.Address) (net.Conn, error) {
	addrs, err := m.resolve(ctx, addr)
	if err != nil {
		m.log.Warnf("failed to resolve %s using rendezvous: %v", addr.String(), err)
		return nil, err
	}

	return m.punch(ctx, addrs, clientConnectionWatcher{}, false)
}

func (m *quicPuncher) resolve(ctx context.Context, addr common.Address) (*sonm.RendezvousReply, error) {
	request := &sonm.ConnectRequest{
		Protocol:     m.protocol,
		PrivateAddrs: []*sonm.Addr{},
		ID:           addr.Bytes(),
	}

	return m.rendezvousClient.Resolve(ctx, request)
}

func (m *quicPuncher) RemoteAddr() net.Addr {
	return m.rendezvousClient.RemoteAddr()
}

func (m *quicPuncher) Close() error {
	return m.rendezvousClient.conn.Close()
}
