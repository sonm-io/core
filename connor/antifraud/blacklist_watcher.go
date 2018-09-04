package antifraud

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	minStep            = time.Hour
	maxStep            = time.Hour * 24 * 7
	unBlacklistTimeout = 30 * time.Second
)

type blacklistWatcher struct {
	log     *zap.Logger
	address common.Address
	client  sonm.BlacklistClient

	nextPeriod    time.Duration
	unBlacklistAt time.Time
	lastSuccess   time.Time
}

func NewBlacklistWatcher(addr common.Address, cc *grpc.ClientConn, log *zap.Logger) *blacklistWatcher {
	return &blacklistWatcher{
		log:        log.Named("blacklist").With(zap.String("wallet", addr.Hex())),
		address:    addr,
		nextPeriod: minStep,
		client:     sonm.NewBlacklistClient(cc),
	}
}

// Failure track task failure by quality.
// Also doubles the next period of blacklisting.
func (m *blacklistWatcher) Failure() {
	m.unBlacklistAt = time.Now().Add(m.nextPeriod)
	m.lastSuccess = time.Time{}

	m.nextPeriod *= 2
	if m.nextPeriod > maxStep {
		m.nextPeriod = maxStep
	}

	m.log.Debug("failure", zap.Duration("step", m.nextPeriod))
}

// Success tracks successful task check is passed,
// and decreases the period of blacklisting by the amount of time
// the task is successfully worked and gives profit.
func (m *blacklistWatcher) Success() {
	if m.lastSuccess.IsZero() {
		m.lastSuccess = time.Now()
		return
	}

	d := time.Now().Sub(m.lastSuccess)
	m.nextPeriod -= d
	if m.nextPeriod < minStep {
		m.nextPeriod = minStep
	}
	m.lastSuccess = time.Now()
}

func (m *blacklistWatcher) isBlacklisted() bool {
	return time.Now().Before(m.unBlacklistAt)
}

func (m *blacklistWatcher) TryUnblacklist(ctx context.Context) error {
	if m.isBlacklisted() || m.unBlacklistAt.IsZero() {
		return nil
	}

	m.log.Info("removing from blacklist on market")
	ctx, cancel := context.WithTimeout(ctx, unBlacklistTimeout)
	defer cancel()
	if _, err := m.client.Remove(ctx, sonm.NewEthAddress(m.address)); err != nil {
		m.log.Warn("cannot remove address from blacklist", zap.Error(err))
		return err
	}

	m.unBlacklistAt = time.Time{}
	return nil

}
