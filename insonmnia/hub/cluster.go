package hub

import (
	"context"
	"encoding/json"
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
	"google.golang.org/grpc"
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
	// Starts synchronization process. Can be called multiple times after EventChannel is closed
	Run() <-chan ClusterEvent

	// IsLeader returns true if this cluster is a leader, i.e. we rule the
	// synchronization process.
	IsLeader() bool

	LeaderClient() (pb.HubClient, error)

	RegisterEntity(name string, prototype interface{})

	Synchronize(entity interface{}) error
}

// Returns a cluster writer interface if this node is a master, event channel
// otherwise.
// Should be recalled when a cluster's master/slave state changes.
// The channel is closed when the specified context is canceled.
func NewCluster(ctx context.Context, cfg *ClusterConfig) (Cluster, error) {
	store, err := makeStore(ctx, cfg)
	if err != nil {
		return nil, err
	}
	endpoints, err := parseEndpoints(cfg)
	if err != nil {
		return nil, err
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
	}
	return &c, nil
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
}

func (c *cluster) Run() <-chan ClusterEvent {
	c.ctx, c.cancel = context.WithCancel(c.parentCtx)
	if c.cfg.Failover {
		c.isLeader = false
		go c.election()
		go c.leaderWatch()
		go c.announce()
		go c.hubWatch()
		go c.hubGC()
	} else {
		log.G(c.ctx).Info("runnning in dev single-server mode")
	}
	go c.watchEvents()
	return c.eventChannel
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

func (c *cluster) RegisterEntity(name string, prototype interface{}) {
	c.registeredEntitiesMu.Lock()
	defer c.registeredEntitiesMu.Unlock()
	t := reflect.TypeOf(prototype)
	c.registeredEntities[name] = t
	c.entityNames[t] = name
}

func (c *cluster) Synchronize(entity interface{}) error {
	if !c.isLeader {
		return errors.New("not a leader")
	}
	name, err := c.nameByEntity(entity)
	if err != nil {
		return err
	}
	data, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	c.store.Put(c.cfg.SynchronizableEntitiesPrefix+"/"+name, data, &store.WriteOptions{})
	return nil
}

func (c *cluster) election() {
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
		case <-c.ctx.Done():
			candidate.Stop()
			return
		}
	}
}

// Blocks in endless cycle watching for leadership.
// When the leadership is changed stores new leader id in cluster
func (c *cluster) leaderWatch() {
	log.G(c.ctx).Info("starting leader watch goroutine")
	follower := leadership.NewFollower(c.store, c.cfg.LeaderKey)
	leaderCh, errCh := follower.FollowElection()
	for {
		select {
		case <-c.ctx.Done():
			follower.Stop()
			return
		case err := <-errCh:
			log.G(c.ctx).Error("leader watch failure", zap.Error(err))
			c.close(errors.WithStack(err))
		case leaderId := <-leaderCh:
			c.leaderLock.Lock()
			c.leaderId = leaderId
			c.leaderLock.Unlock()
			c.emitLeadershipEvent()
		}
	}
}

func (c *cluster) announce() {
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
				return
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *cluster) hubWatch() {
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
				c.close(errors.WithStack(errors.New("hub watcher closed")))
				return
			} else {
				for _, member := range members {
					err := c.registerMember(member)
					if err != nil {
						log.G(c.ctx).Warn("trash data in cluster members folder: ", zap.Any("kvPair", member), zap.Error(err))
					}
				}
			}
		case <-c.ctx.Done():
			close(stopCh)
			return
		}
	}
}

func (c *cluster) checkHub(id string) error {
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

func (c *cluster) hubGC() {
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
			return
		}
	}
}

func (c *cluster) watchEvents() {
	log.G(c.ctx).Info("subscribing on sync folder")
	watchStopChannel := make(chan struct{})
	ch, err := c.store.WatchTree(c.cfg.SynchronizableEntitiesPrefix, watchStopChannel)
	if err != nil {
		c.close(err)
		return
	}
	for {
		select {
		case <-c.ctx.Done():
			close(watchStopChannel)
			return
		case kvList, ok := <-ch:
			if !ok {
				c.close(errors.WithStack(errors.New("watch channel is closed")))
				return
			}
			for _, kv := range kvList {
				name := fetchNameFromPath(kv.Key)
				t, err := c.typeByName(name)
				if err != nil {
					log.G(c.ctx).Warn("unknown synchronizable entity", zap.String("entity", name))
				}
				value := reflect.New(t)
				err = json.Unmarshal(kv.Value, value.Interface())
				if err != nil {
					log.G(c.ctx).Warn("can not unmarshal entity", zap.Error(err))
				} else {
					c.eventChannel <- value.Interface()
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
		return nil, errors.New("entity " + t.String() + " is not registered")
	}
	return t, nil
}

func makeStore(ctx context.Context, cfg *ClusterConfig) (store.Store, error) {
	consul.Register()
	boltdb.Register()
	log.G(ctx).Info("creating store", zap.Any("store", cfg))

	endpoints := []string{cfg.Store.Endpoint}

	backend := store.Backend(cfg.Store.Type)

	config := store.Config{}
	config.Bucket = cfg.Store.Bucket
	return libkv.NewStore(backend, endpoints, &config)
}

func (c *cluster) close(err error) {
	log.G(c.ctx).Error("cluster failure", zap.Error(err))
	c.cancel()
	c.eventChannel <- err
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

func (c *cluster) registerMember(member *store.KVPair) error {
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
	log.G(c.ctx).Info("fetched endpoints of new member", zap.Any("endpoints", endpoints))
	for _, ep := range endpoints {
		conn, err := util.MakeGrpcClient(ep, nil)
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
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.IsGlobalUnicast() {
				endpoints = append(endpoints, ip.String()+":"+port)
			}
		}
	}
	if len(endpoints) == 0 {
		return nil, errors.New("could not determine a single unicast endpoint, check networking")
	}
	return endpoints, nil
}
