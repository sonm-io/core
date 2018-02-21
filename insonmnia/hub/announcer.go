package hub

import (
	"context"
	"crypto/ecdsa"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type Announcer interface {
	// Start starts announcement
	Start(ctx context.Context)
	// Once announces cluster addresses once
	Once(ctx context.Context) error
	// ErrorMsg returns announcement error if any
	ErrorMsg() string
}

type locatorAnnouncer struct {
	mu      sync.Mutex
	key     *ecdsa.PrivateKey
	cluster Cluster
	client  pb.LocatorClient
	period  time.Duration
	err     error
}

func newLocatorAnnouncer(key *ecdsa.PrivateKey, lc pb.LocatorClient, td time.Duration, cls Cluster) Announcer {
	return &locatorAnnouncer{
		key:     key,
		cluster: cls,
		client:  lc,
		period:  td,
	}
}

func (la *locatorAnnouncer) Start(ctx context.Context) {
	tk := util.NewImmediateTicker(la.period)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			if err := la.once(ctx); err != nil {
				log.G(ctx).Warn("cannot announce addresses to Locator", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (la *locatorAnnouncer) Once(ctx context.Context) error {
	return la.once(ctx)
}

func (la *locatorAnnouncer) once(ctx context.Context) error {
	if !la.cluster.IsLeader() {
		return nil
	}

	clientEndpoints, workerEndpoints, err := collectEndpoints(la.cluster)
	if err != nil {
		return err
	}

	req := &pb.AnnounceRequest{
		ClientEndpoints: clientEndpoints,
		WorkerEndpoints: workerEndpoints,
	}

	log.G(ctx).Debug("announcing Hub endpoints",
		zap.String("eth", util.PubKeyToAddr(la.key.PublicKey).Hex()),
		zap.Strings("client_endpoints", clientEndpoints),
		zap.Strings("worker_endpoints", workerEndpoints))

	_, err = la.client.Announce(ctx, req)
	la.keepError(err)
	return err
}

func (la *locatorAnnouncer) keepError(err error) {
	la.mu.Lock()
	defer la.mu.Unlock()

	la.err = err
}

func (la *locatorAnnouncer) ErrorMsg() string {
	la.mu.Lock()
	defer la.mu.Unlock()

	s, ok := status.FromError(la.err)
	if !ok {
		return la.err.Error()
	}

	return s.Message()
}
