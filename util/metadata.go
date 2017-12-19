package util

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// ForwardMetadata is a helper function for gRPC proxy that chains incoming
// request with some outgoing request.
// It simply forwards incoming context metadata without changes by toggling
// internal outgoing key.
func ForwardMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, md)
}
