package util

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MakeGrpcClient creates new gRPC client connection on given addr
// and wraps it with given credentials
func MakeGrpcClient(ctx context.Context, addr string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var secureOpt = grpc.WithInsecure()
	if creds != nil {
		secureOpt = grpc.WithTransportCredentials(creds)
	}
	var extraOpts = append(opts, secureOpt, grpc.WithCompressor(grpc.NewGZIPCompressor()), grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
	cc, err := grpc.DialContext(ctx, addr, extraOpts...)
	if err != nil {
		return nil, err
	}
	return cc, err
}

// MakeGrpcServer creates new gRPC server
func MakeGrpcServer(creds credentials.TransportCredentials, extraopts ...grpc.ServerOption) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.RPCCompressor(grpc.NewGZIPCompressor()),
		grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
	}
	if creds != nil {
		opts = append(opts, grpc.Creds(creds))
	}
	srv := grpc.NewServer(append(opts, extraopts...)...)
	return srv
}
