package node

import (
	"crypto/ecdsa"
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util/rest"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

type Option func(o *options)
type ServerOption func(o *serverOptions) error

type options struct {
	log *zap.Logger
}

func newOptions() *options {
	return &options{
		log: zap.NewNop(),
	}
}

func WithLog(log *zap.Logger) Option {
	return func(o *options) {
		o.log = log
	}
}

type serverOptions struct {
	allowGRPC         bool
	optionsGRPC       []xgrpc.ServerOption
	allowREST         bool
	optionsREST       []rest.Option
	exposeGRPCMetrics bool
	log               *zap.Logger
}

func newServerOptions() *serverOptions {
	return &serverOptions{
		log: zap.NewNop(),
	}
}

func WithGRPCServer() ServerOption {
	return func(o *serverOptions) error {
		o.allowGRPC = true
		return nil
	}
}

func WithGRPCSecure(credentials credentials.TransportCredentials, key *ecdsa.PrivateKey) ServerOption {
	return func(o *serverOptions) error {
		o.optionsGRPC = append(o.optionsGRPC, xgrpc.Credentials(auth.NewWalletAuthenticator(credentials, crypto.PubkeyToAddress(key.PublicKey))))
		return nil
	}
}

func WithRESTServer() ServerOption {
	return func(o *serverOptions) error {
		o.allowREST = true
		return nil
	}
}

func WithRESTSecure(key *ecdsa.PrivateKey) ServerOption {
	return func(o *serverOptions) error {
		hash := sha256.New()
		hash.Write(key.D.Bytes())
		secret := hash.Sum([]byte{})
		codec, err := rest.NewAESDecoderEncoder(secret)
		if err != nil {
			return err
		}

		o.optionsREST = append(o.optionsREST, rest.WithDecoder(codec), rest.WithEncoder(codec))
		return nil
	}
}

func WithGRPCServerMetrics() ServerOption {
	return func(o *serverOptions) error {
		o.exposeGRPCMetrics = true
		return nil
	}
}

func WithServerLog(log *zap.Logger) ServerOption {
	return func(o *serverOptions) error {
		o.log = log
		return nil
	}
}
