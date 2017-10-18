package node

import "google.golang.org/grpc"

// credentials describe any type of auth data.
// Empty for now, but must be implemented in nearest future
type credentials interface{}

// initGrpcClient creates new client connection on addr and wraps it with credentials
func initGrpcClient(addr string, _ credentials) (*grpc.ClientConn, error) {
	// TODO(sshaman1101): crash if cannot connect to addr
	cc, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
	if err != nil {
		return nil, err
	}
	return cc, nil
}
