package hub

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/docker/leadership"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
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

// Specific type of cluster event emited when new member joins cluster
type NewMemberEvent struct {
	Id        string
	endpoints []string
}

// Specific type of cluster event emited when leadership is transferred.
// It is not always loss or aquire of leadership of this specific node
type LeadershipEvent struct {
	Held            bool
	LeaderId        string
	LeaderEndpoints []string
}

type Cluster interface {
	// Starts synchronization process. Can be called multiple times after error is received in EventChannel
	Run() error

	Close()

	// IsLeader returns true if this cluster is a leader, i.e. we rule the
	// synchronization process.
	IsLeader() bool

	LeaderClient() (pb.HubClient, error)

	RegisterAndLoadEntity(name string, prototype interface{}) error

	Synchronize(entity interface{}) error

	// Fetch current cluster members
	Members() ([]NewMemberEvent, error)
}

// Returns a cluster writer interface if this node is a master, event channel
// otherwise.
// Should be recalled when a cluster's master/slave state changes.
// The channel is closed when the specified context is canceled.
func NewCluster(ctx context.Context, cfg *ClusterConfig, creds credentials.TransportCredentials) (Cluster, <-chan ClusterEvent, error) {
	store, err := makeStore(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	endpoints, err := parseEndpoints(cfg)
	if err != nil {
		return nil, nil, err
	}
	c := cluster{
		parentCtx: ctx,
		cfg:       cfg,

		registeredEntities: make(map[string]reflect.Type),
		entityNames:        make(map[reflect.Type]string),

		store: store,

		isLeader:  true,
		id:        uuid.NewV1().String(),
		endpoints: endpoints,

		clients:          make(map[string]*client),
		clusterEndpoints: make(map[string][]string),

		eventChannel: make(chan ClusterEvent, 100),

		creds: creds,
	}

	if cfg.Failover {
		c.isLeader = false
	}

	c.ctx, c.cancel = context.WithCancel(c.parentCtx)
	c.registerMember(c.id, c.endpoints)

	return &c, c.eventChannel, nil
}

type client struct {
	client pb.HubClient
	conn   *grpc.ClientConn
}

type cluster struct {
	parentCtx context.Context
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       *ClusterConfig

	registeredEntitiesMu sync.RWMutex
	registeredEntities   map[string]reflect.Type
	entityNames          map[reflect.Type]string

	store store.Store

	// self info
	isLeader  bool
	id        string
	endpoints []string

	leaderLock sync.RWMutex

	clients          map[string]*client
	clusterEndpoints map[string][]string
	leaderId         string

	eventChannel chan ClusterEvent

	creds credentials.TransportCredentials
}

func (c *cluster) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *cluster) Run() error {
	c.Close()

	w := errgroup.Group{}

	c.ctx, c.cancel = context.WithCancel(c.parentCtx)
	if c.cfg.Failover {
		c.isLeader = false
		w.Go(c.election)
		w.Go(c.leaderWatch)
		w.Go(c.announce)
		w.Go(c.hubWatch)
		w.Go(c.hubGC)
	} else {
		log.G(c.ctx).Info("runnning in dev single-server mode")
	}

	w.Go(c.watchEvents)
	return w.Wait()
}

func (c *cluster) IsLeader() bool {
	return c.isLeader
}

// Get GRPC hub client to current leader
func (c *cluster) LeaderClient() (pb.HubClient, error) {
	log.G(c.ctx).Debug("fetching leader client")
	c.leaderLock.RLock()
	defer c.leaderLock.RUnlock()
	leaderEndpoints, ok := c.clusterEndpoints[c.leaderId]
	if !ok || len(leaderEndpoints) == 0 {
		log.G(c.ctx).Warn("can not determine leader")
		return nil, errors.New("can not determine leader")
	}
	client, ok := c.clients[c.leaderId]
	if !ok || client == nil {
		log.G(c.ctx).Warn("not connected to leader")
		return nil, errors.New("not connected to leader")
	}
	return client.client, nil
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
	if !c.isLeader {
		log.G(c.ctx).Warn("failed to synchronize entity - not a leader")
		return errors.New("not a leader")
	}
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
	log.G(c.ctx).Debug("synchronizing entity", zap.Any("entity", entity), zap.ByteString("marshalled", data))
	c.store.Put(c.cfg.SynchronizableEntitiesPrefix+"/"+name, data, &store.WriteOptions{})
	return nil
}

func (c *cluster) Members() ([]NewMemberEvent, error) {
	result := make([]NewMemberEvent, 0)
	c.leaderLock.RLock()
	defer c.leaderLock.RUnlock()
	for id, endpoints := range c.clusterEndpoints {
		result = append(result, NewMemberEvent{id, endpoints})
	}
	return result, nil
}

func (c *cluster) election() error {
	candidate := leadership.NewCandidate(c.store, c.cfg.LeaderKey, c.id, makeDuration(c.cfg.LeaderTTL))
	electedCh, errCh := candidate.RunForElection()
	log.G(c.ctx).Info("starting leader election goroutine")

	for {
		select {
		case c.isLeader = <-electedCh:
			log.G(c.ctx).Debug("election event", zap.Bool("isLeader", c.isLeader))
			// Do not possibly block on event channel to prevent stale leadership data
			go c.emitLeadershipEvent()
		case err := <-errCh:
			log.G(c.ctx).Error("election failure", zap.Error(err))
			c.close(errors.WithStack(err))
			return err
		case <-c.ctx.Done():
			candidate.Stop()
			return nil
		}
	}
}

// Blocks in endless cycle watching for leadership.
// When the leadership is changed stores new leader id in cluster
func (c *cluster) leaderWatch() error {
	log.G(c.ctx).Info("starting leader watch goroutine")
	follower := leadership.NewFollower(c.store, c.cfg.LeaderKey)
	leaderCh, errCh := follower.FollowElection()
	for {
		select {
		case <-c.ctx.Done():
			follower.Stop()
			return nil
		case err := <-errCh:
			log.G(c.ctx).Error("leader watch failure", zap.Error(err))
			c.close(errors.WithStack(err))
			return err
		case leaderId := <-leaderCh:
			c.leaderLock.Lock()
			c.leaderId = leaderId
			c.leaderLock.Unlock()
			c.emitLeadershipEvent()
		}
	}
}

func (c *cluster) announce() error {
	log.G(c.ctx).Info("starting announce goroutine", zap.Any("endpoints", c.endpoints), zap.String("ID", c.id))
	endpointsData, _ := json.Marshal(c.endpoints)
	ticker := time.NewTicker(makeDuration(c.cfg.AnnounceTTL))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := c.store.Put(c.cfg.MemberListKey+"/"+c.id, endpointsData, &store.WriteOptions{TTL: makeDuration(c.cfg.AnnounceTTL)})
			if err != nil {
				log.G(c.ctx).Error("could not update announce", zap.Error(err))
				c.close(errors.WithStack(err))
				return err
			}
		case <-c.ctx.Done():
			return nil
		}
	}
}

func (c *cluster) hubWatch() error {
	log.G(c.ctx).Info("starting member watch goroutine")
	stopCh := make(chan struct{})
	listener, err := c.store.WatchTree(c.cfg.MemberListKey, stopCh)
	if err != nil {
		c.close(err)
	}
	for {
		select {
		case members, ok := <-listener:
			if !ok {
				err := errors.WithStack(errors.New("hub watcher closed"))
				c.close(err)
				return err
			} else {
				for _, member := range members {
					err := c.registerMemberFromKV(member)
					if err != nil {
						log.G(c.ctx).Warn("trash data in cluster members folder: ", zap.Any("kvPair", member), zap.Error(err))
					}
				}
			}
		case <-c.ctx.Done():
			close(stopCh)
			return nil
		}
	}
}

func (c *cluster) checkHub(id string) error {
	if id == c.id {
		return nil
	}
	exists, err := c.store.Exists(c.cfg.MemberListKey + "/" + id)
	if err != nil {
		return err
	}
	if !exists {
		log.G(c.ctx).Info("hub is offline, removing", zap.String("hubId", id))
		c.leaderLock.Lock()
		defer c.leaderLock.Unlock()
		cli, ok := c.clients[id]
		if ok {
			cli.conn.Close()
			delete(c.clients, id)
		}
	}
	return nil
}

func (c *cluster) hubGC() error {
	log.G(c.ctx).Info("starting hub GC goroutine")
	t := time.NewTicker(makeDuration(c.cfg.MemberGCPeriod))
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.leaderLock.RLock()
			idsToCheck := make([]string, 0)
			for id := range c.clients {
				idsToCheck = append(idsToCheck, id)
			}
			c.leaderLock.RUnlock()

			for _, id := range idsToCheck {
				err := c.checkHub(id)
				if err != nil {
					log.G(c.ctx).Warn("failed to check hub", zap.String("hubId", id), zap.Error(err))
				} else {
					log.G(c.ctx).Info("checked hub", zap.String("hubId", id))
				}
			}

		case <-c.ctx.Done():
			return nil
		}
	}
}

//TODO: extract this to some kind of store wrapper over boltdb
func (c *cluster) watchEventsTree(stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	if c.cfg.Failover {
		return c.store.WatchTree(c.cfg.SynchronizableEntitiesPrefix, stopCh)
	}
	opts := store.WriteOptions{
		IsDir: true,
	}
	empty := make([]byte, 0)
	c.store.Put(c.cfg.SynchronizableEntitiesPrefix, empty, &opts)
	ch := make(chan []*store.KVPair, 1)

	data := make(map[string]*store.KVPair)
	updater := func() error {
		changed := false
		pairs, err := c.store.List(c.cfg.SynchronizableEntitiesPrefix)
		if err != nil {
			return err
		}
		filteredPairs := make([]*store.KVPair, 0)
		for _, pair := range pairs {
			if pair.Key == c.cfg.SynchronizableEntitiesPrefix {
				continue
			}
			filteredPairs = append(filteredPairs, pair)
			cur, ok := data[pair.Key]
			if !ok || !bytes.Equal(cur.Value, pair.Value) {
				changed = true
				data[pair.Key] = pair
			}
		}
		if changed {
			ch <- filteredPairs
		}
		return nil
	}

	if err := updater(); err != nil {
		return nil, err
	}
	go func() {
		t := time.NewTicker(time.Second * 1)
		defer t.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-t.C:
				err := updater()
				if err != nil {
					c.close(err)
				}
			}
		}
	}()
	return ch, nil
}

func (c *cluster) watchEvents() error {
	log.G(c.ctx).Info("subscribing on sync folder")
	watchStopChannel := make(chan struct{})
	ch, err := c.watchEventsTree(watchStopChannel)
	if err != nil {
		c.close(err)
		return err
	}
	for {
		select {
		case <-c.ctx.Done():
			close(watchStopChannel)
			return nil
		case kvList, ok := <-ch:
			if !ok {
				err := errors.WithStack(errors.New("watch channel is closed"))
				c.close(err)
				return err
			}
			for _, kv := range kvList {
				name := fetchNameFromPath(kv.Key)
				t, err := c.typeByName(name)
				if err != nil {
					log.G(c.ctx).Warn("unknown synchronizable entity", zap.String("entity", name))
					continue
				}
				value := reflect.New(t)
				err = json.Unmarshal(kv.Value, value.Interface())
				if err != nil {
					log.G(c.ctx).Warn("can not unmarshal entity", zap.Error(err))
				} else {
					log.G(c.ctx).Debug("received cluster event", zap.String("name", name), zap.Any("value", value.Interface()))
					c.eventChannel <- reflect.Indirect(value).Interface()
				}
			}
		}
	}
}

func (c *cluster) nameByEntity(entity interface{}) (string, error) {
	c.registeredEntitiesMu.RLock()
	defer c.registeredEntitiesMu.RUnlock()
	t := reflect.TypeOf(entity)
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

	endpoints := []string{cfg.Store.Endpoint}

	backend := store.Backend(cfg.Store.Type)

	var tlsConf *tls.Config
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
	return libkv.NewStore(backend, endpoints, &config)
}

func (c *cluster) close(err error) {
	log.G(c.ctx).Error("cluster failure", zap.Error(err))
	c.leaderLock.Lock()
	c.leaderId = ""
	c.isLeader = false
	c.leaderLock.Unlock()
	c.Close()
}

func (c *cluster) emitLeadershipEvent() {
	c.leaderLock.Lock()
	defer c.leaderLock.Unlock()
	endpoints, _ := c.clusterEndpoints[c.leaderId]
	c.eventChannel <- LeadershipEvent{
		Held:            c.isLeader,
		LeaderId:        c.leaderId,
		LeaderEndpoints: endpoints,
	}
}

func (c *cluster) memberExists(id string) bool {
	c.leaderLock.RLock()
	defer c.leaderLock.RUnlock()
	_, ok := c.clients[id]
	return ok
}

func (c *cluster) registerMemberFromKV(member *store.KVPair) error {
	id := fetchNameFromPath(member.Key)
	if id == c.id {
		return nil
	}

	if c.memberExists(id) {
		return nil
	}

	endpoints := make([]string, 0)
	err := json.Unmarshal(member.Value, &endpoints)
	if err != nil {
		return err
	}
	return c.registerMember(id, endpoints)
}

func (c *cluster) registerMember(id string, endpoints []string) error {
	log.G(c.ctx).Info("fetched endpoints of new member", zap.Any("endpoints", endpoints))
	c.leaderLock.Lock()
	c.clusterEndpoints[id] = endpoints
	c.eventChannel <- NewMemberEvent{id, endpoints}
	c.leaderLock.Unlock()

	if id == c.id {
		return nil
	}

	for _, ep := range endpoints {
		conn, err := util.MakeGrpcClient(c.ctx, ep, c.creds)
		if err != nil {
			log.G(c.ctx).Warn("could not connect to hub", zap.String("endpoint", ep), zap.Error(err))
			continue
		} else {
			log.G(c.ctx).Info("successfully connected to cluster member")
			c.leaderLock.Lock()
			defer c.leaderLock.Unlock()
			_, ok := c.clients[id]
			if ok {
				log.G(c.ctx).Info("duplicated connection - dropping")
				conn.Close()
				return nil
			}
			c.clients[id] = &client{pb.NewHubClient(conn), conn}
			return nil
		}
	}
	return errors.New("could not connect to any provided member endpoint")
}

func fetchNameFromPath(key string) string {
	parts := strings.Split(key, "/")
	return parts[len(parts)-1]
}

func makeDuration(numSeconds uint64) time.Duration {
	return time.Second * time.Duration(numSeconds)
}

func parseEndpoints(config *ClusterConfig) ([]string, error) {
	endpoints := make([]string, 0)
	host, port, err := net.SplitHostPort(config.GrpcEndpoint)
	if len(host) != 0 {
		endpoints = append(endpoints, config.GrpcEndpoint)
		return endpoints, nil
	}

	systemIPs, err := util.GetAvailableIPs()
	if err != nil {
		return nil, err
	}

	for _, ip := range systemIPs {
		endpoints = append(endpoints, ip.String()+":"+port)
	}

	return endpoints, nil
}
