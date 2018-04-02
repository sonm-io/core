package hub

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/miner"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

// options for building hub instance
type options struct {
	version   string
	ctx       context.Context
	ethKey    *ecdsa.PrivateKey
	ethAddr   common.Address
	bcr       blockchain.Blockchainer
	market    pb.MarketClient
	locator   pb.LocatorClient
	creds     credentials.TransportCredentials
	rot       util.HitlessCertRotator
	announcer Announcer
	worker    *miner.Miner
}

func defaultHubOptions() *options {
	return &options{
		ctx: context.Background(),
	}
}

// Option func is for applying any params to hub options
type Option func(options *options)

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func WithBlockchain(bcr blockchain.Blockchainer) Option {
	return func(o *options) {
		o.bcr = bcr
	}
}

func WithMarket(mp pb.MarketClient) Option {
	return func(o *options) {
		o.market = mp
	}
}

func WithPrivateKey(k *ecdsa.PrivateKey) Option {
	return func(o *options) {
		o.ethKey = k
		o.ethAddr = util.PubKeyToAddr(k.PublicKey)
	}
}

func WithVersion(v string) Option {
	return func(o *options) {
		o.version = v
	}
}

func WithCreds(creds credentials.TransportCredentials) Option {
	return func(o *options) {
		o.creds = creds
	}
}

func WithCertRotator(rot util.HitlessCertRotator) Option {
	return func(o *options) {
		o.rot = rot
	}
}

func WithWorker(w *miner.Miner) Option {
	return func(o *options) {
		o.worker = w
	}
}
