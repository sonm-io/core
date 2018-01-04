package miner

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
)

func TestGetHubConnectionInfo(t *testing.T) {
	var (
		ctrl                   = gomock.NewController(t)
		loc                    = pb.NewMockLocatorClient(ctrl)
		opts                   = &options{ctx: context.Background(), locatorClient: loc}
		okFullEndpoint         = "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:10002"
		okFullIPv6Endpoint     = "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@[::1]:10002"
		okFullHostnameEndpoint = "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@google.com:10002"
		okResolvedEndpoint     = "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@:80"
		badFullIPv6Endpoint    = "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@::1:10002"
		badResolvedEndpoint    = "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@80"
	)

	loc.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.ResolveReply{IpAddr: []string{"google.com:10001"}}, nil)

	endpoint, err := opts.getHubConnectionInfo(&config{HubConfig: HubConfig{Endpoint: okFullEndpoint}})
	require.Nil(t, err)
	require.Equal(t, endpoint.Endpoint, "127.0.0.1:10002")

	endpoint, err = opts.getHubConnectionInfo(&config{HubConfig: HubConfig{Endpoint: okFullIPv6Endpoint}})
	require.Nil(t, err)
	require.Equal(t, "[::1]:10002", endpoint.Endpoint)

	endpoint, err = opts.getHubConnectionInfo(&config{HubConfig: HubConfig{Endpoint: okFullHostnameEndpoint}})
	require.Nil(t, err)
	require.Equal(t, "google.com:10002", endpoint.Endpoint)

	endpoint, err = opts.getHubConnectionInfo(&config{HubConfig: HubConfig{Endpoint: okResolvedEndpoint}})
	require.Nil(t, err)
	require.Equal(t, "google.com:80", endpoint.Endpoint)

	endpoint, err = opts.getHubConnectionInfo(&config{HubConfig: HubConfig{Endpoint: badFullIPv6Endpoint}})
	require.Error(t, err)

	endpoint, err = opts.getHubConnectionInfo(&config{HubConfig: HubConfig{Endpoint: badResolvedEndpoint}})
	require.Error(t, err)
}
