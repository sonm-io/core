package xgrpc

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
)

func NewServer(logger *zap.Logger, extraOpts ...ServerOption) *grpc.Server {
	opts := newOptions(logger, extraOpts...)
	srv := grpc.NewServer(
		append([]grpc.ServerOption{
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(opts.interceptors.u...)),
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(opts.interceptors.s...)),
		}, opts.options...)...,
	)

	return srv
}

func Services(server *grpc.Server) []string {
	var names []string
	for name := range server.GetServiceInfo() {
		names = append(names, name)
	}

	return names
}
