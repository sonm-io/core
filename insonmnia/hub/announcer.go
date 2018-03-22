package hub

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
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
	// Endpoints which are being announced
	Endpoints() []string
}

type locatorAnnouncer struct {
	mu              sync.Mutex
	key             *ecdsa.PrivateKey
	clientEndpoints []string
	client          pb.LocatorClient
	period          time.Duration
	err             error
}

func newLocatorAnnouncer(key *ecdsa.PrivateKey, lc pb.LocatorClient, td time.Duration, cfg *Config) (Announcer, error) {
	ep, err := getEndpoints(cfg)
	if err != nil {
		return nil, err
	}

	return &locatorAnnouncer{
		key:             key,
		clientEndpoints: ep,
		client:          lc,
		period:          td,
	}, nil
}

func getEndpoints(config *Config) (clientEndpoints []string, err error) {
	clientEndpoint := config.AnnounceEndpoint
	if len(clientEndpoint) == 0 {
		clientEndpoint = config.Endpoint
	}

	clientEndpoints, err = parseEndpoints(clientEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get client endpointsInfo")
	}

	return
}

func parseEndpoints(endpoint string) (endpts []string, err error) {
	ipAddr, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, err
	}

	if len(ipAddr) != 0 {
		ip := net.ParseIP(ipAddr)
		if ip == nil {
			return nil, fmt.Errorf(
				"client endpoint %s must be a valid IP", ipAddr)
		}

		if !ip.IsUnspecified() {
			endpts = append(endpts, endpoint)

			return endpts, nil
		}
	}
	systemIPs, err := util.GetAvailableIPs()
	if err != nil {
		return nil, err
	}

	for _, ip := range systemIPs {
		endpts = append(endpts, net.JoinHostPort(ip.String(), port))
	}

	return endpts, nil
}

func (la *locatorAnnouncer) Endpoints() []string {
	return la.clientEndpoints
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

	req := &pb.AnnounceRequest{
		ClientEndpoints: la.clientEndpoints,
		WorkerEndpoints: []string{},
	}

	log.G(ctx).Debug("announcing Hub endpoints",
		zap.String("eth", util.PubKeyToAddr(la.key.PublicKey).Hex()),
		zap.Strings("client_endpoints", la.clientEndpoints),
		zap.Strings("worker_endpoints", []string{}))

	_, err := la.client.Announce(ctx, req)
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
