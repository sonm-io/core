package xgrpc

import (
	"crypto/tls"

	. "github.com/sonm-io/core/util"
	"google.golang.org/grpc/credentials"
)

// TransportCredentials wraps the standard transport credentials, adding an
// ability to obtain the TLS config used.
type TransportCredentials struct {
	credentials.TransportCredentials
	TLSConfig *tls.Config
}

func NewTransportCredentials(cfg *tls.Config) *TransportCredentials {
	return &TransportCredentials{
		TransportCredentials: NewTLS(cfg),
		TLSConfig:            cfg,
	}
}
