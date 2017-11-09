package hub

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/docker/leadership"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/satori/uuid"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"
)

const leaderKey = "sonm/hub/leader"
const listKey = "sonm/hub/list"
const synchronizableEntitiesPrefix = "sonm/hub/sync"

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
func NewCluster(ctx context.Context, cfg *ClusterConfig) (Cluster, <-chan ClusterEvent, error) {
	store, err := makeStore(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	endpoints, err := parseEndpoints(cfg)
	if err != nil {
		return nil, nil, err
	}
	c := cluster{
		parentCtx:        ctx,
		cfg:              cfg,
		store:            store,
		isLeader:         true,
		id:               uuid.NewV1().String(),
		endpoints:        endpoints,
		clients:          make(map[string]pb.HubClient),
		clusterEndpoints: make(map[string][]string),
		eventChannel:     make(chan ClusterEvent, 100),
	}
	return &c, c.eventChannel, nil
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

	leaderLock sync.Mutex

	clients          map[string]pb.HubClient
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
	}
	go c.watchEvents()
	return c.eventChannel
}

func (c *cluster) IsLeader() bool {
	return c.isLeader
}

// Get GRPC hub client to current leader
func (c *cluster) LeaderClient() (pb.HubClient, error) {
	c.leaderLock.Lock()
	defer c.leaderLock.Unlock()
	leaderEndpoints, ok := c.clusterEndpoints[c.leaderId]
	if !ok || len(leaderEndpoints) == 0 {
		return nil, errors.New("can not determine leader")
	}
	client, ok := c.clients[c.leaderId]
	if !ok || client == nil {
		return nil, errors.New("not connected to leader")
	}
	return client, nil
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
	c.store.Put(synchronizableEntitiesPrefix+"/"+name, data, &store.WriteOptions{})
	return nil
}

func (c *cluster) election() {
	candidate := leadership.NewCandidate(c.store, leaderKey, c.id, 5*time.Second)
	electedCh, errCh := candidate.RunForElection()
	log.G(c.ctx).Info("starting leader election goroutine")

	for {
		select {
		case c.isLeader = <-electedCh:
			c.emitLeadershipEvent()
		case err := <-errCh:
			c.close(err)
		case <-c.ctx.Done():
			return
		}
	}
}

// Blocks in endless cycle watching for leadership.
// When the leadership is changed stores new leader id in cluster
func (c *cluster) leaderWatch() {
	follower := leadership.NewFollower(c.store, leaderKey)
	leaderCh, errCh := follower.FollowElection()
	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-errCh:
			c.close(err)
		case c.leaderId = <-leaderCh:
			c.emitLeadershipEvent()
		}
	}
}

func (c *cluster) announce() {
	endpointsData, _ := json.Marshal(c.endpoints)
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			err := c.store.Put(listKey+"/"+c.id, endpointsData, &store.WriteOptions{TTL: time.Second * 5})
			if err != nil {
				c.close(err)
				return
			}
		case <-c.ctx.Done():
			return
		}
	}

}

func (c *cluster) hubWatch() {
	stopCh := make(chan struct{})
	listener, err := c.store.WatchTree(listKey, stopCh)
	if err != nil {
		c.close(err)
	}
	for {
		select {
		case members, ok := <-listener:
			if !ok {
				c.close(errors.New("hub watcher closed"))
				return
			} else {
				for _, member := range members {
					err := c.registerMember(member)
					if err != nil {
						log.G(c.ctx).Warn("trash data in cluster members folder: ", zap.Any("kvPair", member))
					}
				}
			}
		case <-c.ctx.Done():
			stopCh <- struct{}{}
			return
		}
	}
}

func (c *cluster) watchEvents() {
	watchStopChannel := make(chan struct{})
	ch, err := c.store.WatchTree(synchronizableEntitiesPrefix, watchStopChannel)
	if err != nil {
		c.close(err)
		return
	}
	for {
		select {
		case <-c.ctx.Done():
		case kvList, ok := <-ch:
			if !ok {
				c.close(errors.New("watch channel is closed"))
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
					c.eventChannel <- value.Interface()
				}
			}
		}
	}
}

func (c *cluster) nameByEntity(entity interface{}) (string, error) {
	c.registeredEntitiesMu.RLock()
	t := reflect.TypeOf(entity)
	name, ok := c.entityNames[t]
	if !ok {
		return "", errors.New("entity " + t.String() + " is not registered")
	}
	return name, nil
}

func (c *cluster) typeByName(name string) (reflect.Type, error) {
	c.registeredEntitiesMu.RLock()
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

	endpoints := []string{cfg.StoreEndpoint}

	backend := store.Backend(cfg.StoreType)

	config := store.Config{}
	config.Bucket = cfg.StoreBucket
	return libkv.NewStore(backend, endpoints, &config)
}

func (c *cluster) close(err error) {
	log.G(c.ctx).Error("cluster failure", zap.Error(err))
	c.eventChannel <- err
	close(c.eventChannel)
}

func (c *cluster) emitLeadershipEvent() {
	c.leaderLock.Lock()
	endpoints, _ := c.clusterEndpoints[c.leaderId]
	c.eventChannel <- LeadershipEvent{
		Held:            c.isLeader,
		LeaderId:        c.leaderId,
		LeaderEndpoints: endpoints,
	}
	c.leaderLock.Unlock()
}

func (c *cluster) registerMember(member *store.KVPair) error {
	id := fetchNameFromPath(member.Key)

	endpoints := make([]string, 0)
	err := json.Unmarshal(member.Value, endpoints)
	if err != nil {
		return err
	}
	for _, ep := range endpoints {
		conn, err := util.MakeGrpcClient(ep, nil)
		if err != nil {
			log.G(c.ctx).Warn("could not connect to hub", zap.String("endpoint", ep), zap.Error(err))
			continue
		} else {
			c.leaderLock.Lock()
			c.clients[id] = pb.NewHubClient(conn)
			c.leaderLock.Unlock()
			return nil
		}
	}
	return errors.New("could not connect to any provided member endpoint")
}

func fetchNameFromPath(key string) string {
	parts := strings.Split(key, "/")
	return parts[len(parts)-1]
}

func parseEndpoints(config *ClusterConfig) ([]string, error) {
	endpoints := make([]string, 0)
	if len(config.GrpcIp) != 0 {
		return append(endpoints, config.GrpcIp+":"+string(config.GrpcPort)), nil
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
				endpoints = append(endpoints, ip.String())
			}
		}
	}
	if len(endpoints) == 0 {
		return nil, errors.New("could not determine a single unicast endpoint, check networking")
	}
	return endpoints, nil
}
