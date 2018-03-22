package relay

import (
	"fmt"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/sonm-io/core/proto"
)

// Dial communicates with the relay server waiting for other server peer to
// establish a relayed TCP connection.
//
// All network traffic will be transported through that server.
func Dial(addr net.Addr, targetAddr common.Address, uuid string) (net.Conn, error) {
	client, err := newClient(addr)
	if err != nil {
		return nil, err
	}

	defer client.Close()

	member, err := client.discover(targetAddr)
	if err != nil {
		return nil, err
	}

	return member.dial(targetAddr, uuid)
}

// Listen publishes itself to the relay server waiting for other client peer
// to establish a relayed TCP connection.
//
// All network traffic will be transported through that server.
func Listen(addr net.Addr, publishAddr SignedETHAddr) (net.Conn, error) {
	client, err := newClient(addr)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	member, err := client.discover(publishAddr.addr)
	if err != nil {
		return nil, err
	}

	return member.accept(publishAddr)
}

type client struct {
	conn net.Conn
}

func newClient(addr net.Addr) (*client, error) {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	m := &client{
		conn: conn,
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

	return newClient(addr)
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
