package locator

import (
	"crypto/ecdsa"
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

func NewApp(conf *Config, key *ecdsa.PrivateKey) (*App, error) {
	if key == nil {
		return nil, errors.New("private key should be provided")
	}

	logger := logging.BuildLogger(-1, true)
	ctx := ctxlog.WithLogger(context.Background(), logger)

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, errors.Wrap(err, "canot init CertRotator")
	}

	creds := util.NewTLS(TLSConfig)

	app := &App{
		conf:   conf,
		ethKey: key,

		creds:       creds,
		grpc:        util.MakeGrpcServer(creds),
		certRotator: certRotator,
	}

	pb.RegisterLocatorServer(app.grpc,
		NewServer(inmemory.NewStorage(conf.CleanupPeriod, conf.NodeTTL, ctxlog.GetLogger(ctx))))

	return app, nil
}
