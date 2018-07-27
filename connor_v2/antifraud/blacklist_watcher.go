package antifraud

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
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
}

func (m *blacklistWatcher) Failure() {
	m.till = time.Now().Add(m.currentStep)
	m.currentStep *= 2
	if m.currentStep > maxStep {
		m.currentStep = maxStep
	}
}

func (m *blacklistWatcher) Success() {
	m.currentStep /= minStep
	if m.currentStep < minStep {
		m.currentStep = minStep
	}
}

func (m *blacklistWatcher) Blacklisted() bool {
	return time.Now().Before(m.till)
}

func (m *blacklistWatcher) TryUnblacklist(ctx context.Context) error {
	if m.Blacklisted() {
		return nil
	}
	_, err := m.client.Remove(ctx, sonm.NewEthAddress(m.address))
	return err

}
