package xgrpc

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewServer(logger *zap.Logger, extraOpts ...ServerOption) *grpc.Server {
	opts := newOptions(logger, extraOpts...)
	srv := grpc.NewServer(
		append([]grpc.ServerOption{
			grpc.RPCCompressor(grpc.NewGZIPCompressor()),
			grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(opts.interceptors.u...)),
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(opts.interceptors.s...)),
		}, opts.options...)...,
	)

	return srv
}
