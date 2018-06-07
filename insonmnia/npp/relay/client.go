package relay

import (
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
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

	member, err := client.discover(targetAddr)
	if err != nil {
		log.Warn("failed to discover meeting point on the Continuum", zap.Error(err))
		return nil, err
	}

	log.Debug("connecting to remote meeting point on the Continuum", zap.Stringer("remote_addr", member.conn.RemoteAddr()))
	conn, err := member.dial(targetAddr, uuid)
	if err != nil {
		log.Warn("failed to connect to remote meeting point on the Continuum", zap.Error(err))
		return nil, err
	}

	return conn, err
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

	conn, err := net.Dial("tcp", addr.String())
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
