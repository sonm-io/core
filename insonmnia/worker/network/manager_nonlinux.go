// +build !linux

package network

import (
	"errors"

	"github.com/sonm-io/core/insonmnia/worker/network/tc"
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
)

type TC struct{}

func (TC) QDiscAdd(qdisc tc.QDisc) error {
	return ErrUnsupportedPlatform
}

func (TC) QDiscDel(qdisc tc.QDisc) error {
	return ErrUnsupportedPlatform
}

func (TC) ClassAdd(class tc.Class) error {
	return ErrUnsupportedPlatform
}

func (TC) ClassDel(class tc.Class) error {
	return ErrUnsupportedPlatform
}

func (TC) FilterAdd(filter tc.Filter) error {
	return ErrUnsupportedPlatform
}

func (TC) FilterDel(filter tc.Filter) error {
	return ErrUnsupportedPlatform
}

func (TC) Close() error {
	return nil
}

func (m *localNetworkManager) Init() error {
	m.tc = &TC{}
	return nil
}

func (m *localNetworkManager) NewActions(network *Network) []Action {
	return []Action{
		&DockerNetworkCreateAction{
			DockerClient: m.dockerClient,
			Network:      network,
		},
	}
}
