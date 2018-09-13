package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type TincNetworkState struct {
	ctx      context.Context
	config   *TincNetworkConfig
	mu       sync.RWMutex
	cli      *client.Client
	Networks map[string]*TincNetwork
	Pools    map[string]*net.IPNet
	logger   *zap.SugaredLogger
	storage  store.Store
}

func defaultNet() *net.IPNet {
	_, r, _ := net.ParseCIDR("10.20.30.0/24")
	return r
}

func newTincNetworkState(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkState, error) {
	storage, err := makeStore(ctx, config.StatePath)
	if err != nil {
		return nil, err
	}

	state := &TincNetworkState{
		ctx:      ctx,
		config:   config,
		cli:      client,
		Networks: map[string]*TincNetwork{},
		storage:  storage,
		logger:   log.S(ctx).With("source", "tinc/state"),
	}

	err = state.load()
	if err != nil {
		return nil, err
	}
	return state, nil
}

func getNetByCIDR(cidr string) (*net.IPNet, error) {
	if len(cidr) != 0 {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		return ipNet, nil
	} else {
		return defaultNet(), nil
	}
}

func (t *TincNetworkState) InsertTincNetwork(n *sonm.NetworkSpec, cgroupParent string) (*TincNetwork, error) {
	pool, err := getNetByCIDR(n.Subnet)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(n.Addr)
	if ip.Mask(pool.Mask).Equal(pool.IP) {
		return nil, errors.New("ip does not match network pool")
	}

	containerConfig := &container.Config{
		Image: "sonm/tinc",
	}
	hostConfig := &container.HostConfig{
		Privileged:  true,
		NetworkMode: "host",
		Resources: container.Resources{
			CgroupParent: cgroupParent,
		},
		AutoRemove: true,
	}
	netConfig := &network.NetworkingConfig{}

	resp, err := t.cli.ContainerCreate(t.ctx, containerConfig, hostConfig, netConfig, "")
	if err != nil {
		t.logger.Errorf("failed to create tinc container: %s", err)
		return nil, err
	}
	err = t.cli.ContainerStart(t.ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		t.logger.Errorf("failed to start tinc container: %s", err)
		return nil, err
	}
	log.S(t.ctx).Infof("started container %s", resp.ID)

	_, enableBridge := n.Options["enable_bridge"]
	invitation, _ := n.Options["invitation"]

	result := &TincNetwork{
		NodeID:          n.GetID(),
		DockerID:        "",
		Pool:            pool,
		Invitation:      invitation,
		EnableBridge:    enableBridge,
		CgroupParent:    cgroupParent,
		ConfigPath:      t.config.ConfigDir + "/" + n.GetID(),
		TincContainerID: resp.ID,
		cli:             t.cli,
		logger:          t.logger.With("source", "tinc/network/"+n.GetID(), "container", resp.ID),
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Networks[result.NodeID] = result
	return result, nil
}

func (t *TincNetworkState) netByID(id string) (*TincNetwork, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, ok := t.Networks[id]
	if !ok {
		return nil, fmt.Errorf("could not find network by id %s", id)
	}
	return n, nil
}
func (t *TincNetworkState) netByOptions(data map[string]interface{}) (*TincNetwork, error) {
	var id interface{}
	id, ok := data["id"]
	if !ok {
		g, ok := data["com.docker.network.generic"]
		if ok {
			id, _ = g.(map[string]interface{})["id"]
		}
	}

	if id == nil {
		return nil, errors.New("missing id in option is required")
	}
	return t.netByID(id.(string))
}

func (t *TincNetworkState) netByIPAMOptions(data map[string]string) (*TincNetwork, error) {
	id, ok := data["id"]
	if !ok {
		t.logger.Warnw("missing id field in options", zap.Any("options", data))
		return nil, errors.New("missing id field in options")
	}
	return t.netByID(id)
}

func (t *TincNetworkState) netByDockerID(id string) (*TincNetwork, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, n := range t.Networks {
		if n.DockerID == id {
			return n, nil
		}
	}
	return nil, fmt.Errorf("network not found by docker id %s", id)
}

func makeStore(ctx context.Context, path string) (store.Store, error) {
	boltdb.Register()
	s := store.Backend(store.BOLTDB)
	endpoints := []string{path}
	config := store.Config{
		Bucket: "sonm_tinc_driver_state",
	}
	return libkv.NewStore(s, endpoints, &config)
}

func (t *TincNetworkState) load() (err error) {
	defer func() {
		if err == store.ErrKeyNotFound {
			err = nil
		}
		if err != nil {
			t.logger.Errorf("could not load tinc network state - %s; erasing key", err)
			delErr := t.storage.Delete("state")
			if delErr != nil {
				t.logger.Errorf("could not cleanup storage for tinc network: %s", delErr)
			}
		}
	}()

	exists, err := t.storage.Exists("state")
	if err != nil || !exists {
		return
	}

	data, err := t.storage.Get("state")
	if err != nil {
		return
	}

	err = json.Unmarshal(data.Value, t)
	if err != nil {
		return
	}
	for _, n := range t.Networks {
		n.cli = t.cli
		n.logger = t.logger.With("source", "tinc/network/"+n.NodeID, "container", n.TincContainerID)
	}
	return
}

func (t *TincNetworkState) sync() error {
	var err error
	defer func() {
		if err != nil {
			t.logger.Errorf("could not sync network state: %s", err)
		}
	}()

	marshalled, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = t.storage.Put("state", marshalled, &store.WriteOptions{})
	return err
}
