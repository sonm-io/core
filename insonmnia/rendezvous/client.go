// This package contains a tiny wrapper over the generated gRPC rendezvous
// client that allows to close the client explicitly.

package rendezvous

import (
	"context"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client extends the generated gRPC client allowing to close the underlying
// connection.
// Users should call Client.Close to terminate all the pending operations.
type Client interface {
	sonm.RendezvousClient
	// Close closes this client freeing the associated resources.
	//
	// All pending operations will be terminated with error.
	Close() error
}

type client struct {
	sonm.RendezvousClient
	conn *grpc.ClientConn
}

// NewRendezvousClient constructs a new rendezvous client.
//
// The address provided will be used for establishing a TCP connection while
// optional credentials - for authentication. Additionally other dial options
// can be specified.
func NewRendezvousClient(ctx context.Context, addr string, credentials credentials.TransportCredentials, opts ...grpc.DialOption) (Client, error) {
	conn, err := xgrpc.NewWalletAuthenticatedClient(ctx, credentials, addr, opts...)
	if err != nil {
		return nil, err
	}

	return &client{sonm.NewRendezvousClient(conn), conn}, nil
}

func (m *client) Close() error {
	return m.conn.Close()
}
