package network

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sonm-io/core/insonmnia/worker/network/tc"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
)

const (
	networkPrefix       = "sonm"
	networkIfbPrefix    = "ifb-"
	driverBridge        = "bridge"
	tagSonmNetwork      = "com.sonm.network"
	tagSonmNetworkAlias = "com.sonm.network.alias"
	maxLinkNameLen      = 12
)

// Action is an abstraction of some action that can be rolled back.
type Action interface {
	// Execute executes this action, returning error if any.
	Execute(ctx context.Context) error
	// Rollback rollbacks this action, returning error if any.
	Rollback() error
}

// ActionQueue represents a queue of executable actions.
// Any action that fails triggers cascade previous actions rollback.
type ActionQueue struct {
	vec []Action
	mu  sync.Mutex
}

func NewActionQueue(actions ...Action) *ActionQueue {
	return &ActionQueue{
		vec: actions,
	}
}

// Execute executes the given action.
//
// If that action fails a cascade previous actions rollback occurs resulting
// in a tuple of this's action error and rollback ones if any.
func (m *ActionQueue) Execute(ctx context.Context, action Action) (error, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := action.Execute(ctx); err != nil {
		return err, m.rollback()
	}

	m.vec = append(m.vec, action)

	return nil, nil
}

func (m *ActionQueue) Rollback() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.rollback()
}

func (m *ActionQueue) rollback() error {
	errs := multierror.NewMultiError()
	for {
		if action, ok := m.pop(); ok {
			errs = multierror.Append(errs, action.Rollback())
		} else {
			break
		}
	}

	return errs.ErrorOrNil()
}

func (m *ActionQueue) pop() (Action, bool) {
	length := len(m.vec)
	if length == 0 {
		return nil, false
	}

	action := m.vec[length-1]
	m.vec = m.vec[:length-1]
	return action, true
}

type Network struct {
	ID               string
	Name             string
	Alias            string
	RateLimitEgress  uint64
	RateLimitIngress uint64
}

type CreateNetworkRequest struct {
	// ID specifies network interface suffix.
	ID               string
	RateLimitEgress  uint64
	RateLimitIngress uint64
}

type PruneRequest struct {
	// IDs specify network interface suffixes.
	IDs []string
}

type PruneReply struct {
	// Result contains information about network removing result.
	Result map[string]error
}

type NetworkManagerConfig struct {
	DockerClient *client.Client
	Log          *zap.SugaredLogger
}

type NetworkManager struct {
	dockerClient *client.Client
	tc           tc.TC
	log          *zap.SugaredLogger
}

// NewNetworkManager constructs a new network manager.
//
// Some basic checks are performed during execution of this function, like
// checking whether the host OS is capable to limit network bandwidth etc.
func NewNetworkManager() (*NetworkManager, error) {
	return NewNetworkManagerWithConfig(NetworkManagerConfig{})
}

func NewNetworkManagerWithConfig(cfg NetworkManagerConfig) (*NetworkManager, error) {
	var err error
	if cfg.DockerClient == nil {
		cfg.DockerClient, err = client.NewEnvClient()
		if err != nil {
			return nil, err
		}
	}
	if cfg.Log == nil {
		cfg.Log = zap.NewNop().Sugar()
	}

	m := &NetworkManager{
		dockerClient: cfg.DockerClient,
		log:          cfg.Log,
	}

	if err := m.init(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *NetworkManager) CreateNetwork(ctx context.Context, request *CreateNetworkRequest) (*Network, error) {
	name := fmt.Sprintf("%s%s", networkPrefix, request.ID)
	nameLink, ok := truncLinkName(name)
	if ok {
		m.log.Debugf("truncated network name %s -> %s", name, nameLink)
	}

	network := &Network{
		Name:             nameLink,
		Alias:            name,
		RateLimitEgress:  request.RateLimitEgress,
		RateLimitIngress: request.RateLimitIngress,
	}

	actionQueue := NewActionQueue()

	for _, action := range m.newActions(network) {
		err, errs := actionQueue.Execute(ctx, action)
		if err != nil {
			m.log.Errorw("failed to setup network", zap.Error(err))
			if errs != nil {
				m.log.Errorw("failed to rollback network", zap.Error(errs))
			}

			return nil, err
		}
	}

	return network, nil
}

func (m *NetworkManager) RemoveNetwork(network *Network) error {
	return NewActionQueue(m.newActions(network)...).Rollback()
}

// Prune tries to remove all unused networks that look like a SONM networks.
//
// We attach a special tag to Docker networks, so we can identify a network
// created by us from others.
func (m *NetworkManager) Prune(ctx context.Context, request *PruneRequest) (*PruneReply, error) {
	networkSet := map[string]struct{}{}
	for _, ID := range request.IDs {
		name := fmt.Sprintf("%s%s", networkPrefix, ID)
		nameLink, ok := truncLinkName(name)
		if ok {
			m.log.Debugf("truncated network name %s -> %s", name, nameLink)
		}

		networkSet[nameLink] = struct{}{}
	}

	result := map[string]error{}

	filter := filters.NewArgs()
	filter.Add("label", tagSonmNetwork)
	networks, err := m.dockerClient.NetworkList(ctx, types.NetworkListOptions{
		Filters: filter,
	})
	if err != nil {
		return nil, err
	}

	m.log.Debugw("found SONM networks", zap.Any("networks", networks))

	for _, network := range networks {
		result[network.Name] = m.dockerClient.NetworkRemove(ctx, network.ID)
	}

	return &PruneReply{Result: result}, nil
}

func (m *NetworkManager) Close() error {
	return m.tc.Close()
}

type DockerNetworkCreateAction struct {
	DockerClient *client.Client
	Network      *Network
}

func (m *DockerNetworkCreateAction) Execute(ctx context.Context) error {
	network, err := m.DockerClient.NetworkCreate(ctx, m.Network.Name, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         driverBridge,
		Options: map[string]string{
			"com.docker.network.bridge.name": m.Network.Name,
		},
		Labels: map[string]string{
			tagSonmNetwork:      "",
			tagSonmNetworkAlias: m.Network.Alias,
		},
	})
	m.Network.ID = network.ID

	if m.isErrNetworkAlreadyExists(err) {
		return nil
	}

	return err
}

func (m *DockerNetworkCreateAction) Rollback() error {
	if m.Network.ID == "" {
		return fmt.Errorf("no network was created before")
	}
	return m.DockerClient.NetworkRemove(context.Background(), m.Network.ID)
}

func (m *DockerNetworkCreateAction) isErrNetworkAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	// Providing errors API? Nevermind, we'll check your errors by finding
	// substrings.
	return strings.Contains(err.Error(), fmt.Sprintf("network with name %s already exists", m.Network.Name))
}

// TruncLinkName truncates network interface name to match OS requirements,
// returning truncated name with flag pointing at whether is was truncated.
//
// In recent kernel versions this is defined by IFNAMSIZ constant. Also have
// in mind that if this interface is planned to be used with DHCP, dhclient
// does not properly support interface names with length > 13.
func truncLinkName(v string) (string, bool) {
	if len(v) > maxLinkNameLen {
		return v[:maxLinkNameLen], true
	}
	return v, false
}
