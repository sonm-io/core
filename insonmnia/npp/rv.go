package npp

import (
	"context"
	"net"

	"fmt"

	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc/credentials"
)

// RendezvousClient is a tiny wrapper over the generated gRPC client allowing
// to close the underlying connection and to get its connection info that is
// useful for NAT penetration.
type rendezvousClient struct {
	rendezvous.Client

	// The underlying connection. Held here for information reasons, it is
	// closed internally in the gRPC client.
	conn net.Conn
}

func newRendezvousClient(ctx context.Context, addr auth.Addr, creds credentials.TransportCredentials) (*rendezvousClient, error) {
	targetAddr, err := addr.ETH()
	if err != nil {
		return nil, err
	}

	netAddr, err := addr.Addr()
	if err != nil {
		return nil, err
	}

	client, conn, err := connectToRendezvous(ctx, addr, creds)
	if err != nil {
		return nil, err
	}

	resp, err := client.Discover(ctx, &sonm.HandshakeRequest{PeerType: sonm.PeerType_CLIENT, Addr: targetAddr.Bytes()})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to discover rendezvous node: %s", err)
	}

	if resp.Addr != netAddr {
		conn.Close()
		client, conn, err := connectToRendezvous(ctx, auth.NewAddrRaw(targetAddr, resp.Addr), creds)
		if err != nil {
			return nil, err
		}
		return &rendezvousClient{client, conn}, nil
	}

	return &rendezvousClient{client, conn}, nil
}

func connectToRendezvous(ctx context.Context, addr auth.Addr, creds credentials.TransportCredentials) (rendezvous.Client, net.Conn, error) {
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
		return nil, nil, err
	}

	conn, err := dialer.Dial(protocol, netAddr)
	if err != nil {
		return nil, nil, err
	}

	client, err := rendezvous.NewRendezvousClient(ctx, addr.String(), creds, xgrpc.WithConn(conn))
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	return client, conn, nil
}

// LocalAddr returns the local network address.
func (m *rendezvousClient) LocalAddr() net.Addr {
	return m.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (m *rendezvousClient) RemoteAddr() net.Addr {
	return m.conn.RemoteAddr()
}
