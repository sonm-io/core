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
	"strings"
	"sync"
	"time"
)

const leaderKey = "sonm/hub/leader"
const listKey = "sonm/hub/list"

// ClusterEvent describes an event that can produce the cluster.
//
// Possible types are:
// - `map[string]DeviceProperties` when received device properties updates.
// - `T` types for other synchronizable entities.
// - `struct{}` when switched the state.
// - `error` when a connection to the Consul is broken.
// TODO: maybe add some typed errors?

type ClusterEvent interface{}

type Cluster interface {
	// IsLeader returns true if this cluster is a leader, i.e. we rule the
	// synchronization process.
	IsLeader() bool

	LeaderClient() (pb.HubClient, error)
	//TryForwardToLeader(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo) (bool, interface{}, error)

	// All these operations should fail if this node is not a leader.

	SynchronizeTasks(map[string]*TaskInfo) error
	// SynchronizeDevices synchronizes device properties with followers.
	//SynchronizeDevices(properties map[string]DeviceProperties) error

	// ... and so on for other stuff we need to synchronize.
	//SynchronizeTasks(...)
	//SynchronizeSlots(...)
	//SynchronizeACL(...)
}

type cluster struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *ClusterConfig

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

func (c *cluster) SynchronizeTasks(map[string]*TaskInfo) error {
	return nil
}

// Returns a cluster writer interface if this node is a master, event channel
// otherwise.
// Should be recalled when a cluster's master/slave state changes.
// The channel is closed when the specified context is canceled.
func NewCluster(ctx context.Context, cfg *ClusterConfig) (Cluster, <-chan ClusterEvent, error) {
	ctx, cancel := context.WithCancel(ctx)
	store, err := makeStore(ctx, cfg)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	endpoints, err := parseEndpoints(cfg)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	c := cluster{
		ctx:       ctx,
		cfg:       cfg,
		cancel:    cancel,
		id:        uuid.NewV1().String(),
		endpoints: endpoints,
		store:     store,
		isLeader:  true,
	}
	if cfg.Failover {
		c.isLeader = false
		go c.election()
	}
	return &c, c.eventChannel, nil
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

func (c *cluster) election() {
	go c.leaderWatch()
	go c.hubWatch()

	candidate := leadership.NewCandidate(c.store, leaderKey, c.id, 5*time.Second)
	electedCh, errCh := candidate.RunForElection()
	log.G(c.ctx).Info("starting leader election goroutine")

	for {
		select {
		case c.isLeader = <-electedCh:
		case err := <-errCh:
			c.close(err)
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
		}
	}
}

func (c *cluster) hubWatch() error {
	// TODO: can this ever fail?
	endpointsData, _ := json.Marshal(c.endpoints)

	go func() {
		ticker := time.NewTicker(time.Second * 1)
		select {
		case <-ticker.C:
			err := c.store.Put(listKey+"/"+c.id, endpointsData, &store.WriteOptions{TTL: time.Second * 5})
			if err != nil {
				c.close(err)
			}
		case <-c.ctx.Done():
			return
		}
	}()

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
		}
	}
}

func (c *cluster) registerMember(member *store.KVPair) error {
	id := fetchIdFromKey(member.Key)

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

func fetchIdFromKey(key string) string {
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
