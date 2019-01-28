package network

import (
	"context"
	"time"

	"github.com/docker/docker/client"
	"github.com/sonm-io/core/proto"
)

var _ networkManager = &remoteNetworkManager{}

type RemoteNetworkAliasAction struct {
	Client  sonm.QOSClient
	Network *Network
}

func (m *RemoteNetworkAliasAction) Execute(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.Client.SetAlias(ctx, &sonm.QOSSetAliasRequest{
		LinkName:  m.Network.Name,
		LinkAlias: m.Network.Alias,
	})

	return err
}

func (m *RemoteNetworkAliasAction) Rollback() error {
	// The alias will be removed with the associated interface, so do nothing
	// here.
	return nil
}

type RemoteHTBShapingAction struct {
	Client  sonm.QOSClient
	Network *Network
}

func (m *RemoteHTBShapingAction) Execute(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := m.Client.AddHTBShaping(ctx, &sonm.QOSAddHTBShapingRequest{
		LinkName:         m.Network.Name,
		LinkAlias:        m.Network.Alias,
		RateLimitEgress:  m.Network.RateLimitEgress,
		RateLimitIngress: m.Network.RateLimitIngress,
	})

	return err
}

func (m *RemoteHTBShapingAction) Rollback() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.Client.RemoveHTBShaping(ctx, &sonm.QOSRemoveHTBShapingRequest{
		LinkName: m.Network.Name,
	})

	return err
}

type remoteNetworkManager struct {
	client       sonm.QOSClient
	dockerClient *client.Client
}

func (m *remoteNetworkManager) Init() error {
	return nil
}

func (m *remoteNetworkManager) Close() error {
	return nil
}

func (m *remoteNetworkManager) NewActions(network *Network) []Action {
	return []Action{
		&DockerNetworkCreateAction{
			DockerClient: m.dockerClient,
			Network:      network,
		},
		&RemoteNetworkAliasAction{
			Client:  m.client,
			Network: network,
		},
		&RemoteHTBShapingAction{
			Client:  m.client,
			Network: network,
		},
	}
}
