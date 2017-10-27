package util

import "google.golang.org/grpc"

// Credentials describe any type of auth data.
// Empty for now, but must be implemented in nearest future
type Credentials interface{}

// MakeGrpcClient creates new gRPC client connection on given addr
// and wraps it with given credentials
func MakeGrpcClient(addr string, _ Credentials) (*grpc.ClientConn, error) {
	cc, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
	if err != nil {
		return nil, err
	}
	return cc, err
}

// MakeGrpcServer creates new gRPC server
func MakeGrpcServer() *grpc.Server {
	srv := grpc.NewServer(
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
	)
	return srv
}
