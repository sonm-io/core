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
	minStep = time.Hour
	maxStep = time.Hour * 24 * 7
)

type blacklistWatcher struct {
	address     common.Address
	till        time.Time
	currentStep time.Duration
	client      sonm.BlacklistClient
	log         *zap.Logger
}

func NewBlacklistWatcher(addr common.Address, cc *grpc.ClientConn, log *zap.Logger) *blacklistWatcher {
	return &blacklistWatcher{
		log:         log.Named("blacklist").With(zap.String("wallet", addr.Hex())),
		address:     addr,
		currentStep: minStep,
		client:      sonm.NewBlacklistClient(cc),
	}
}

func (m *blacklistWatcher) Failure() {
	m.till = time.Now().Add(m.currentStep)
	m.currentStep *= 2
	if m.currentStep > maxStep {
		m.currentStep = maxStep
	}
	m.log.Debug("failure", zap.Duration("step", m.currentStep))
}

func (m *blacklistWatcher) Success() {
	m.currentStep /= 2
	if m.currentStep < minStep {
		m.currentStep = minStep
	}
	m.log.Debug("success", zap.Duration("step", m.currentStep))
}

func (m *blacklistWatcher) Blacklisted() bool {
	return time.Now().Before(m.till)
}

func (m *blacklistWatcher) TryUnblacklist(ctx context.Context) error {
	if m.Blacklisted() {
		return nil
	}

	m.till = time.Time{}

	m.log.Info("removing from blacklist on market")
	_, err := m.client.Remove(ctx, sonm.NewEthAddress(m.address))
	return err

}
