package locator

import (
	"crypto/ecdsa"
	"net"
	"sync"
	"time"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"

	engine "github.com/sonm-io/core/insonmnia/locator/storage/libkv"
	"github.com/sonm-io/core/insonmnia/logging"
)

func init() {
	boltdb.Register()
}

type App struct {
	mx sync.Mutex

	conf   *Config
	ethKey *ecdsa.PrivateKey

	grpc        *grpc.Server
	certRotator util.HitlessCertRotator
	creds       credentials.TransportCredentials
}

func (a *App) Serve() error {
	lis, err := net.Listen("tcp", a.conf.ListenAddr)
	if err != nil {
		return err
	}

	return a.grpc.Serve(lis)
}

func NewApp(conf *Config, key *ecdsa.PrivateKey) *App {
	return &App{conf: conf, ethKey: key}
}

func (a *App) Init() error {
	if a.conf == nil {
		return errors.New("conf option cannot be nil")
	}

	if a.ethKey == nil {
		return errors.New("private key should be provided")
	}

	// init logger
	logger := logging.BuildLogger(-1, true)
	ctx := ctxlog.WithLogger(context.Background(), logger)

	// init cert rotator
	cr, TLSConfig, err := util.NewHitlessCertRotator(ctx, a.ethKey)
	if err != nil {
		return errors.Wrap(err, "cannot init CertRotator")
	}
	a.certRotator = cr

	// init storage
	logger.Info("initializing storage", zap.Any("storage", a.conf.Store))

	kv, err := libkv.NewStore(
		store.Backend(a.conf.Store.Type),
		[]string{a.conf.Store.Endpoint},
		&store.Config{
			Bucket:            a.conf.Store.Bucket,
			ConnectionTimeout: 10 * time.Second,
		},
	)
	if err != nil {
		logger.Fatal("cannot initialize storage engine", zap.Error(err))
	}

	storage, err := engine.NewStorage(a.conf.NodeTTL, kv)
	if err != nil {
		logger.Fatal("cannot initialize storage", zap.Error(err))
	}

	// init gRPC server
	a.creds = util.NewTLS(TLSConfig)
	a.grpc = util.MakeGrpcServer(a.creds)
	pb.RegisterLocatorServer(a.grpc, NewServer(storage))

	return nil
}
