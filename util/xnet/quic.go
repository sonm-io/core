package xnet

import (
	"crypto/tls"
	"net"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/qerr"
	"github.com/sonm-io/core/util/multierror"
)

type quicError struct {
	error
}

func newQUICError(err error) *quicError {
	return &quicError{
		error: err,
	}
}

func (m *quicError) Timeout() bool {
	if err := qerr.ToQuicError(m.error); err != nil {
		return err.Timeout()
	}

	return false
}

func (m *quicError) Temporary() bool {
	return m.Timeout()
}

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

func (m *QUICConn) Close() error {
	errs := multierror.NewMultiError()

	if err := m.Stream.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := m.session.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
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
	for {
		session, err := m.Listener.Accept()
		if err != nil {
			return nil, newQUICError(err)
		}

		stream, err := session.AcceptStream()
		if err != nil {
			if isPeerGoneErr(err) {
				continue
			}

			return nil, err
		}

		conn := &QUICConn{
			Stream:  stream,
			session: session,
		}

		return conn, nil
	}
}

func isPeerGoneErr(err error) bool {
	if qErr, ok := err.(*qerr.QuicError); ok && qErr.ErrorCode != qerr.PeerGoingAway {
		return true
	}

	return false
}
