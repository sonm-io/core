package network

import (
	"context"
	"encoding/json"
	"net"
	"sync"

	"github.com/docker/docker/client"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type TincNetworkState struct {
	ctx             context.Context
	config          *TincNetworkConfig
	mu              sync.RWMutex
	cli             *client.Client
	Networks        map[string]*TincNetwork
	networkNameToId map[string]string
	Pools           map[string]*net.IPNet
	logger          *zap.SugaredLogger
	storage         store.Store
}

func newTincNetworkState(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkState, error) {
	storage, err := makeStore(ctx, config.StatePath)
	if err != nil {
		return nil, err
	}

	state := &TincNetworkState{
		ctx:             ctx,
		config:          config,
		cli:             client,
		Networks:        map[string]*TincNetwork{},
		networkNameToId: map[string]string{},
		Pools:           map[string]*net.IPNet{},
		storage:         storage,
		logger:          log.S(ctx).With("source", "tinc/state"),
	}

	err = state.load()
	if err != nil {
		return nil, err
	}
	return state, nil
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

func (t *TincNetworkState) RegisterNetworkMapping(id string, name string) error {
	if len(name) == 0 || len(id) == 0 {
		return errors.Errorf("invalid network mapping arguments: \"%s\" \"%s\"", id, name)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.Networks[id]
	if !ok {
		return errors.Errorf("no network with id %s", id)
	}
	t.networkNameToId[name] = id
	return nil
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
				t.logger.Errorf("could not cleanup storage for tinc network - %s", delErr)
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
		n.logger = t.logger.With("source", "tinc/network/"+n.ID, "container", n.TincContainerID)
		t.RegisterNetworkMapping(n.ID, n.Name)
	}
	return
}

func (t *TincNetworkState) sync() error {
	var err error
	defer func() {
		if err != nil {
			t.logger.Errorf("could not sync network state - %s", err)
		}
	}()

	marshalled, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = t.storage.Put("state", marshalled, &store.WriteOptions{})
	return err
}
