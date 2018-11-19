package npp

import (
	"context"
	"net"

	"github.com/libp2p/go-reuseport"
	"github.com/lucas-clemente/quic-go"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/sonm-io/core/util/xnet"
	"google.golang.org/grpc/credentials"
)

// RendezvousClient is a tiny wrapper over the generated gRPC client allowing
// to close the underlying connection and to get its connection info that is
// useful for NAT penetration.
type rendezvousClient struct {
	rendezvous.Client

	// The underlying connection. Held here for information reasons, it is
	// closed internally in the gRPC client.
	conn    net.Conn
	UDPConn net.PacketConn
}

func newRendezvousClient(ctx context.Context, addr auth.Addr, credentials credentials.TransportCredentials) (*rendezvousClient, error) {
	// Setting TCP keepalive is required, because NAT's conntrack can purge out
	// idle connections for its internal garbage collection reasons at the most
	// inopportune moment.
	dialer := reuseport.Dialer{
		D: net.Dialer{
			KeepAlive: tcpKeepAliveInterval,
		},
	}

	netAddr, err := addr.Addr()
	if err != nil {
		return nil, err
	}

	conn, err := dialer.Dial(protocol, netAddr)
	if err != nil {
		return nil, err
	}

	client, err := rendezvous.NewRendezvousClient(ctx, addr.String(), credentials, xgrpc.WithConn(conn))
	if err != nil {
		return nil, err
	}

	return &rendezvousClient{Client: client, conn: conn}, nil
}

func newRendezvousQUICClient(ctx context.Context, udpConn net.PacketConn, addr auth.Addr, credentials *xgrpc.TransportCredentials) (*rendezvousClient, error) {
	netAddr, err := addr.Addr()
	if err != nil {
		return nil, err
	}

	udpNetAddr, err := net.ResolveUDPAddr("udp", netAddr)
	if err != nil {
		return nil, err
	}

	session, err := quic.Dial(udpConn, udpNetAddr, netAddr, credentials.TLSConfig, xnet.DefaultQUICConfig())
	if err != nil {
		return nil, err
	}

	conn, err := xnet.NewQUICConn(session)
	if err != nil {
		return nil, err
	}

	client, err := rendezvous.NewRendezvousClient(ctx, addr.String(), credentials, xgrpc.WithConn(conn))
	if err != nil {
		return nil, err
	}

	return &rendezvousClient{Client: client, conn: conn, UDPConn: udpConn}, nil
}

// LocalAddr returns the local network address.
func (m *rendezvousClient) LocalAddr() net.Addr {
	return m.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (m *rendezvousClient) RemoteAddr() net.Addr {
	return m.conn.RemoteAddr()
}
