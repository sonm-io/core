// This module is responsible for client-side Relay communication.

package relay

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/netutil"
	"go.uber.org/zap"
)

const (
	tcpKeepAliveInterval = 15 * time.Second
)

// Dial communicates with the relay server waiting for other server peer to
// establish a relayed TCP connection.
//
// All network traffic will be transported through that server.
func Dial(addr net.Addr, targetAddr common.Address, uuid string) (net.Conn, error) {
	return DialWithLog(addr, targetAddr, uuid, zap.NewNop())
}

// DialWithLog does the same as Dial, but with logging.
func DialWithLog(addr net.Addr, targetAddr common.Address, uuid string, log *zap.Logger) (net.Conn, error) {
	client, err := newClient(addr, log)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	log = log.With(zap.Stringer("addr", targetAddr))
	log.Debug("discovering meeting point on the Continuum")

	for numAttempts := 0; numAttempts < 2; numAttempts++ {
		member, err := client.discover(targetAddr)
		if err != nil {
			log.Warn("failed to discover meeting point on the Continuum", zap.Error(err))
			return nil, err
		}

		log.Debug("connecting to remote meeting point on the Continuum", zap.Stringer("remote_addr", member.conn.RemoteAddr()))
		conn, err := member.dial(targetAddr, uuid)
		if err == nil {
			return conn, nil
		}

		if verboseErr, ok := err.(*protocolError); ok && verboseErr.code == ErrWrongNode {
			continue
		} else {
			log.Warn("failed to connect to remote meeting point on the Continuum", zap.Error(err))
			return nil, err
		}
	}

	return nil, errors.New("failed to dial remote")
}

// Listen publishes itself to the relay server waiting for other client peer
// to establish a relayed TCP connection.
//
// All network traffic will be transported through that server.
func Listen(addr net.Addr, publishAddr SignedETHAddr) (net.Conn, error) {
	return ListenWithLog(addr, publishAddr, zap.NewNop())
}

// ListenWithLog does the same as Listen, but with logging.
func ListenWithLog(addr net.Addr, publishAddr SignedETHAddr, log *zap.Logger) (net.Conn, error) {
	client, err := newClient(addr, log)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	log = log.With(zap.Stringer("addr", publishAddr.Addr()))
	log.Debug("discovering meeting point on the Continuum")

	member, err := client.discover(publishAddr.addr)
	if err != nil {
		log.Warn("failed to discover meeting point on the Continuum", zap.Error(err))
		return nil, err
	}

	log.Debug("listening for connections on remote meeting point on the Continuum", zap.Stringer("remote_addr", member.conn.RemoteAddr()))
	conn, err := member.accept(publishAddr)
	if err != nil {
		log.Warn("failed to accept connection on remote meeting point on the Continuum", zap.Error(err))
		return nil, err
	}

	return conn, err
}

type client struct {
	conn net.Conn
	log  *zap.Logger
}

func newClient(addr net.Addr, log *zap.Logger) (*client, error) {
	log = log.With(zap.Stringer("remote_addr", addr))
	log.Debug("connecting to the Relay")

	// Setting TCP keepalive is strictly suggested, because in case of mobile
	// devices the network can be reconfigured multiple times between
	// connection attempts.
	// However if the network was silently changed or the connection was lost
	// or the other side was kernel-panicked no FIN/RST will be delivered to
	// us, which leads to infinite (well, 24-hour) hanging.
	dialer := net.Dialer{
		KeepAlive: tcpKeepAliveInterval,
	}

	conn, err := dialer.Dial("tcp", addr.String())
	if err != nil {
		log.Warn("failed to connect to the Relay", zap.Error(err))
		return nil, err
	}

	m := &client{
		conn: conn,
		log:  log,
	}

	return m, nil
}

func (m *client) discover(peer common.Address) (*client, error) {
	if err := sendFrame(m.conn, newDiscover(peer)); err != nil {
		return nil, err
	}

	response := &sonm.DiscoverResponse{}
	if err := recvFrame(m.conn, response); err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", response.Addr)
	if err != nil {
		return nil, err
	}

	return newClient(addr, m.log)
}

func (m *client) dial(targetAddr common.Address, uuid string) (net.Conn, error) {
	if err := m.handshake(newClientHandshake(targetAddr, uuid)); err != nil {
		m.conn.Close()
		return nil, err
	}

	return m.conn, nil
}

func (m *client) accept(publishAddr SignedETHAddr) (net.Conn, error) {
	if err := m.handshake(newServerHandshake(publishAddr)); err != nil {
		m.conn.Close()
		return nil, err
	}

	return m.conn, nil
}

func (m *client) handshake(message proto.Message) error {
	if err := sendFrame(m.conn, message); err != nil {
		return err
	}

	response := &sonm.HandshakeResponse{}
	if err := recvFrame(m.conn, response); err != nil {
		return err
	}

	if response.Error != 0 {
		return fmt.Errorf("failed to perform handshake into relay: %s", response.Description)
	}

	return nil
}

func (m *client) Close() error {
	return m.conn.Close()
}

// Dialer is a thing that constructs a new TCP connection using remote Relay
// server.
// One or more Relay TCP addresses must be specified in "Addrs" field.
// Hostname resolution is performed for each of them for environments with
// dynamic DNS addition/removal. Thus, a single Relay endpoint as a hostname
// should fit the best.
type Dialer struct {
	Addrs []string
	Log   *zap.Logger
}

// Dial mimics "net.Dial" and connects to a remote endpoint using Relay server.
func (m *Dialer) Dial(target common.Address) (net.Conn, error) {
	m.initLog()
	m.Log.Debug("connecting to remote Relay server")

	errs := multierror.NewMultiError()

	for _, addr := range m.Addrs {
		m.Log.Debug("resolving Relay addr", zap.String("addr", addr))

		addrs, err := netutil.LookupTCPHostPort(addr)
		if err != nil {
			errs = multierror.AppendUnique(errs, err)
			continue
		}

		m.Log.Debug("successfully resolved Relay addr", zap.String("addr", addr), zap.Any("resolved", addrs))

		for _, addr := range addrs {
			conn, err := DialWithLog(addr, target, "", m.Log)
			if err == nil {
				return conn, nil
			}

			errs = multierror.AppendUnique(errs, err)
		}
	}

	return nil, fmt.Errorf("failed to connect to %+v: %s", m.Addrs, errs.Error())
}

func (m *Dialer) initLog() {
	if m.Log == nil {
		m.Log = zap.NewNop()
	}
}

// Listener represents client-side TCP listener that will route traffic through
// Relay server.
// One or more Relay TCP addresses must be specified in "Addrs" field.
// Hostname resolution is performed for each of them for environments with
// dynamic DNS addition/removal.
type Listener struct {
	Addrs      []string
	SignedAddr SignedETHAddr
	Log        *zap.Logger
}

func NewListener(addrs []string, key *ecdsa.PrivateKey, log *zap.Logger) (*Listener, error) {
	signedAddr, err := NewSignedAddr(key)
	if err != nil {
		return nil, err
	}

	m := &Listener{
		Addrs:      addrs,
		SignedAddr: signedAddr,
		Log:        log.With(zap.Any("addrs", addrs)),
	}

	return m, nil
}

// Accept accepts a new TCP connection that is relayed through the Relay server
// specified in the configuration.
//
// This is done by establishing a TCP connection to a Relay server and waiting
// until a remote peer decides to communicate with us.
//
// This method can be called multiple times, even concurrently to announce the
// server multiple times, which is useful for reducing the time window between
// the next announce after the remote peer appearing. However, this increases
// resource usage, such as fd, on both this server and the Relay.
// Use wisely.
func (m *Listener) Accept() (net.Conn, error) {
	m.Log.Debug("connecting to remote Relay server")

	errs := multierror.NewMultiError()
	for _, addr := range m.Addrs {
		m.Log.Debug("resolving Relay addr", zap.String("addr", addr))
		addrs, err := netutil.LookupTCPHostPort(addr)
		if err != nil {
			errs = multierror.AppendUnique(errs, err)
			continue
		}

		m.Log.Debug("successfully resolved Relay addr", zap.String("addr", addr), zap.Any("resolved", addrs))

		for _, addr := range addrs {
			conn, err := ListenWithLog(addr, m.SignedAddr, m.Log)
			if err == nil {
				return conn, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to %+v: %s", m.Addrs, errs.Error())
}
