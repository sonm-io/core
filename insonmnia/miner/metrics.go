package miner

import (
	"expvar"

	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

var (
	statsMap = expvar.NewMap("miner_stats")

	bytesSentToHub       expvar.Int
	bytesReceivedFromHub expvar.Int
	uncompressedRPCBytes expvar.Int
)

func init() {
	statsMap.Set("bytes_sent_to_hub", &bytesSentToHub)
	statsMap.Set("bytes_received_from_hub", &bytesReceivedFromHub)
	statsMap.Set("uncompressed_rpc_bytes", &uncompressedRPCBytes)
}

type grpcStatsHandler struct{}

func (grpcStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (grpcStatsHandler) HandleRPC(_ context.Context, s stats.RPCStats) {
	// Count compressed ingress traffic from hub connection
	// Count compressed outgress and raw rpc bytes to estimate compression efficiency
	switch s := s.(type) {
	case *stats.InHeader:
		bytesReceivedFromHub.Add(int64(s.WireLength))
	case *stats.InPayload:
		bytesReceivedFromHub.Add(int64(s.WireLength))
	case *stats.InTrailer:
		bytesReceivedFromHub.Add(int64(s.WireLength))
	case *stats.OutHeader:
		bytesSentToHub.Add(int64(s.WireLength))
	case *stats.OutPayload:
		bytesSentToHub.Add(int64(s.WireLength))
		uncompressedRPCBytes.Add(int64(s.Length))
	case *stats.OutTrailer:
		bytesSentToHub.Add(int64(s.WireLength))
	}
}

func (grpcStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context { return ctx }

func (grpcStatsHandler) HandleConn(context.Context, stats.ConnStats) {}
