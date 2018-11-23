package xnet

import (
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"
)

func DefaultQUICConfig() *quic.Config {
	return &quic.Config{
		Versions: []quic.VersionNumber{
			quic.VersionGQUIC39,
			quic.VersionGQUIC43,
			quic.VersionMilestone0_10_0,
		},
		KeepAlive: true,
	}
}

type QUICConn struct {
	quic.Stream
	session quic.Session
}

func NewQUICConn(session quic.Session) (*QUICConn, error) {
	stream, err := session.OpenStream()
	if err != nil {
		return nil, err
	}

	m := &QUICConn{
		Stream:  stream,
		session: session,
	}

	return m, nil
}

func (m *QUICConn) LocalAddr() net.Addr {
	return m.session.LocalAddr()
}

func (m *QUICConn) RemoteAddr() net.Addr {
	return m.session.RemoteAddr()
}

func ListenQUIC(network, address string, tlsConfig *tls.Config, config *quic.Config) (*QUICListener, error) {
	conn, err := net.ListenPacket(network, address)
	if err != nil {
		return nil, err
	}

	listener, err := quic.Listen(conn, tlsConfig, config)
	if err != nil {
		return nil, err
	}

	return &QUICListener{Listener: listener}, nil
}

type QUICListener struct {
	quic.Listener
}

func (m *QUICListener) Accept() (net.Conn, error) {
	session, err := m.Listener.Accept()
	if err != nil {
		return nil, err
	}

	stream, err := session.AcceptStream()
	if err != nil {
		return nil, err
	}

	conn := &QUICConn{
		Stream:  stream,
		session: session,
	}

	return conn, nil
}
