// +build !linux

package network

import (
	"context"
	"errors"

	"github.com/sonm-io/core/insonmnia/worker/network/tc"
	"github.com/sonm-io/core/proto"
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

type nilQOS struct{}

func (nilQOS) SetAlias(context.Context, *sonm.QOSSetAliasRequest) (*sonm.QOSSetAliasResponse, error) {
	return nil, ErrUnsupportedPlatform
}
func (nilQOS) AddHTBShaping(context.Context, *sonm.QOSAddHTBShapingRequest) (*sonm.QOSAddHTBShapingResponse, error) {
	return nil, ErrUnsupportedPlatform
}
func (nilQOS) RemoveHTBShaping(context.Context, *sonm.QOSRemoveHTBShapingRequest) (*sonm.QOSRemoveHTBShapingResponse, error) {
	return nil, ErrUnsupportedPlatform
}
func (nilQOS) Flush(context.Context, *sonm.QOSFlushRequest) (*sonm.QOSFlushResponse, error) {
	return nil, ErrUnsupportedPlatform
}

func NewRemoteQOS() (sonm.QOSServer, error) {
	return &nilQOS{}, nil
}
