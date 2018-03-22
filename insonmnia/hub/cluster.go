package hub

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// ClusterEvent describes an event that can produce the cluster.
//
// Possible types are:
// - `NewMemberEvent` when new member joins cluster
// - `LeadershipEvent` when leadership is transferred
// - `T` types for other registered synchronizable entities.
// - `error` on any unrecoverable error, after that channel is closed
//   and the user should call Run once more to enable synchronization

type ClusterEvent interface{}

type Cluster interface {
	RegisterAndLoadEntity(name string, prototype interface{}) error
	Synchronize(entity interface{}) error
}

// Returns a cluster writer interface if this node is a master, event channel
// otherwise.
// Should be recalled when a cluster's master/slave state changes.
// The channel is closed when the specified context is canceled.
func NewCluster(ctx context.Context, cfg *ClusterConfig, creds credentials.TransportCredentials) (Cluster, error) {
	clusterStore, err := makeStore(ctx, cfg)
	if err != nil {
		return nil, err
	}

	err = clusterStore.Put(cfg.SynchronizableEntitiesPrefix, []byte{}, &store.WriteOptions{IsDir: true})
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	c := cluster{
		ctx:                ctx,
		cfg:                cfg,
		registeredEntities: make(map[string]reflect.Type),
		entityNames:        make(map[reflect.Type]string),
		store:              clusterStore,
		creds:              creds,
	}

	return &c, nil
}

type cluster struct {
	ctx context.Context
	cfg *ClusterConfig

	registeredEntitiesMu sync.RWMutex
	registeredEntities   map[string]reflect.Type
	entityNames          map[reflect.Type]string

	store store.Store

	leaderLock sync.RWMutex

	creds credentials.TransportCredentials
}

func (c *cluster) RegisterAndLoadEntity(name string, prototype interface{}) error {
	c.registeredEntitiesMu.Lock()
	defer c.registeredEntitiesMu.Unlock()
	t := reflect.Indirect(reflect.ValueOf(prototype)).Type()
	c.registeredEntities[name] = t
	c.entityNames[t] = name
	keyName := c.cfg.SynchronizableEntitiesPrefix + "/" + name
	exists, err := c.store.Exists(keyName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not check entity %s for existance in storage", name))
	}
	if !exists {
		return nil
	}
	kvPair, err := c.store.Get(keyName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not fetch entity %s initial value from storage", name))
	}
	err = json.Unmarshal(kvPair.Value, prototype)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not unmarshal entity %s from storage data", name))
	}
	return nil
}

func (c *cluster) Synchronize(entity interface{}) error {
	name, err := c.nameByEntity(entity)
	if err != nil {
		log.G(c.ctx).Warn("unknown synchronizable entity", zap.Any("entity", entity))
		return err
	}
	data, err := json.Marshal(entity)
	if err != nil {
		log.G(c.ctx).Warn("could not marshal entity", zap.Error(err))
		return err
	}
	log.G(c.ctx).Debug("synchronizing entity", zap.Any("entity", entity), zap.ByteString("marshaled", data))
	c.store.Put(c.cfg.SynchronizableEntitiesPrefix+"/"+name, data, &store.WriteOptions{})
	return nil
}

func (c *cluster) nameByEntity(entity interface{}) (string, error) {
	c.registeredEntitiesMu.RLock()
	defer c.registeredEntitiesMu.RUnlock()
	t := reflect.Indirect(reflect.ValueOf(entity)).Type()
	name, ok := c.entityNames[t]
	if !ok {
		return "", errors.New("entity " + t.String() + " is not registered")
	}
	return name, nil
}

func (c *cluster) typeByName(name string) (reflect.Type, error) {
	c.registeredEntitiesMu.RLock()
	defer c.registeredEntitiesMu.RUnlock()
	t, ok := c.registeredEntities[name]
	if !ok {
		return nil, errors.New("entity " + name + " is not registered")
	}
	return t, nil
}

func makeStore(ctx context.Context, cfg *ClusterConfig) (store.Store, error) {
	consul.Register()
	boltdb.Register()
	log.G(ctx).Info("creating store", zap.Any("store", cfg))

	var (
		endpts  = []string{cfg.Store.Endpoint}
		backend = store.Backend(cfg.Store.Type)
		tlsConf *tls.Config
	)
	if len(cfg.Store.CertFile) != 0 && len(cfg.Store.KeyFile) != 0 {
		cer, err := tls.LoadX509KeyPair(cfg.Store.CertFile, cfg.Store.KeyFile)
		if err != nil {
			return nil, err
		}

		tlsConf = &tls.Config{
			Certificates: []tls.Certificate{cer},
		}
	}
	config := store.Config{
		TLS: tlsConf,
	}
	config.Bucket = cfg.Store.Bucket

	return libkv.NewStore(backend, endpts, &config)
}
