package util

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MakeGrpcClient creates new gRPC client connection on given addr
// and wraps it with given credentials
func MakeGrpcClient(addr string, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	var secureOpt = grpc.WithInsecure()
	if creds != nil {
		secureOpt = grpc.WithTransportCredentials(creds)
	}
	cc, err := grpc.Dial(addr, secureOpt,
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
	if err != nil {
		return nil, err
	}
	return cc, err
}

// MakeGrpcServer creates new gRPC server
func MakeGrpcServer(creds credentials.TransportCredentials) *grpc.Server {
	var opts = []grpc.ServerOption{grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor())}
	if creds != nil {
		opts = append(opts, grpc.Creds(creds))
	}
	srv := grpc.NewServer(opts...)
	return srv
}
