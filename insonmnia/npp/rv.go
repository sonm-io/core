package npp

import (
	"context"
	"net"

	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/rendezvous"
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

func newRendezvousClient(ctx context.Context, addr auth.Endpoint, credentials credentials.TransportCredentials) (*rendezvousClient, error) {
	// Setting TCP keepalive is required, because NAT's conntrack can purge out
	// idle connections for its internal garbage collection reasons at the most
	// inopportune moment.
	dialer := reuseport.Dialer{
		D: net.Dialer{
			KeepAlive: tcpKeepAliveInterval,
		},
	}
	conn, err := dialer.Dial(protocol, addr.Endpoint)
	if err != nil {
		return nil, err
	}

	client, err := rendezvous.NewRendezvousClient(ctx, "", credentials, xgrpc.WithConn(conn))
	if err != nil {
		return nil, err
	}

	return &rendezvousClient{client, conn}, nil
}

// LocalAddr returns the local network address.
func (m *rendezvousClient) LocalAddr() net.Addr {
	return m.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (m *rendezvousClient) RemoteAddr() net.Addr {
	return m.conn.RemoteAddr()
}
