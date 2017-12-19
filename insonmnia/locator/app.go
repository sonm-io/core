package locator

import (
	"crypto/ecdsa"
	"crypto/tls"
	"net"
	"sync"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"

	"github.com/sonm-io/core/insonmnia/locator/storage/inmemory"
	"github.com/sonm-io/core/insonmnia/logging"
)

type App struct {
	mx sync.Mutex

	conf   *Config
	ethKey *ecdsa.PrivateKey

	grpc        *grpc.Server
	certRotator util.HitlessCertRotator
	creds       credentials.TransportCredentials
}

func (l *App) Serve() error {
	lis, err := net.Listen("tcp", l.conf.ListenAddr)
	if err != nil {
		return err
	}

	return l.grpc.Serve(lis)
}

func NewApp(conf *Config, key *ecdsa.PrivateKey) (l *App, err error) {

	if key == nil {
		return nil, errors.Wrap(err, "private key should be provided")
	}

	logger := logging.BuildLogger(-1, true)
	ctx := ctxlog.WithLogger(context.Background(), logger)

	l = &App{
		conf:   conf,
		ethKey: key,
	}

	var TLSConfig *tls.Config
	l.certRotator, TLSConfig, err = util.NewHitlessCertRotator(ctx, l.ethKey)
	if err != nil {
		return nil, err
	}

	l.creds = util.NewTLS(TLSConfig)
	l.grpc = util.MakeGrpcServer(l.creds)

	pb.RegisterLocatorServer(l.grpc,
		NewServer(inmemory.NewStorage(conf.CleanupPeriod, conf.NodeTTL, ctxlog.GetLogger(ctx))))

	return l, nil
}
