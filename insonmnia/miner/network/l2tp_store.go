package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type l2tpState struct {
	Aliases  map[string]string
	Networks map[string]*l2tpNetwork
	mu       sync.Mutex
	ctx      context.Context
	storage  store.Store
}

func newL2TPNetworkStore(ctx context.Context, path string) (*l2tpState, error) {
	boltdb.Register()
	var (
		backend   = store.Backend(store.BOLTDB)
		endpoints = []string{path}
		config    = store.Config{Bucket: "sonm_l2tp_driver_state"}
	)
	storage, err := libkv.NewStore(backend, endpoints, &config)
	if err != nil {
		return nil, err
	}

	state := &l2tpState{
		Aliases:  make(map[string]string),
		Networks: make(map[string]*l2tpNetwork),
		ctx:      ctx,
		storage:  storage,
	}

	if err := state.load(); err != nil {
		return nil, err
	}

	return state, nil
}

func (s *l2tpState) Lock() {
	s.mu.Lock()
}

func (s *l2tpState) Unlock() {
	s.mu.Unlock()
}

func (s *l2tpState) Sync() (err error) {
	defer func() {
		if err != nil {
			log.G(s.ctx).Error("failed to sync l2tp state", zap.Error(err))
		}
	}()

	marshaled, err := json.Marshal(s)
	if err != nil {
		return err
	}

	err = s.storage.Put("state", marshaled, &store.WriteOptions{})
	return
}

func (s *l2tpState) AddNetwork(netID string, netInfo *l2tpNetwork) error {
	if _, ok := s.Aliases[netID]; ok {
		return errors.Errorf("network already exists: %s", netID)
	}

	s.Aliases[netID] = netID
	s.Networks[netID] = netInfo

	return nil
}

func (s *l2tpState) AddNetworkAlias(netID, alias string) error {
	if _, ok := s.Aliases[netID]; !ok {
		return errors.Errorf("network not found: %s", netID)
	}

	s.Aliases[alias] = netID

	return nil
}

func (s *l2tpState) RemoveNetwork(netID string) error {
	translatedID, ok := s.Aliases[netID]
	if !ok {
		return errors.Errorf("network not found: %s", netID)
	}

	delete(s.Networks, translatedID)

	for alias, target := range s.Aliases {
		if target == translatedID {
			delete(s.Aliases, alias)
		}
	}

	return nil
}

func (s *l2tpState) GetNetwork(netID string) (*l2tpNetwork, error) {
	translatedID, ok := s.Aliases[netID]
	if !ok {
		return nil, errors.Errorf("network not found: %s", netID)
	}

	if netInfo, ok := s.Networks[translatedID]; ok {
		return netInfo, nil
	}

	return nil, errors.Errorf("network not found: %s", netID)
}

func (s *l2tpState) GetNetworks() []*l2tpNetwork {
	var out []*l2tpNetwork
	for _, netInfo := range s.Networks {
		out = append(out, netInfo)
	}

	return out
}

type l2tpNetwork struct {
	ID          string
	PoolID      string
	Count       int
	NetworkOpts *L2TPNetworkConfig
	Store       *l2tpEndpointStore
}

func newL2tpNetwork(opts *L2TPNetworkConfig) *l2tpNetwork {
	return &l2tpNetwork{
		NetworkOpts: opts,
		Store:       NewL2TPEndpointStore(),
	}
}

func (n *l2tpNetwork) Setup() error {
	n.PoolID = n.NetworkOpts.GetHash()

	return nil
}

func (n *l2tpNetwork) ConnInc() {
	n.Count++
}

type l2tpEndpointStore struct {
	Aliases   map[string]string
	Endpoints map[string]*l2tpEndpoint
}

func NewL2TPEndpointStore() *l2tpEndpointStore {
	return &l2tpEndpointStore{
		Aliases:   make(map[string]string),
		Endpoints: make(map[string]*l2tpEndpoint),
	}
}

func (s *l2tpEndpointStore) AddEndpoint(endpointID string, eptInfo *l2tpEndpoint) error {
	if _, ok := s.Aliases[endpointID]; ok {
		return errors.Errorf("endpoint already exists: %s", endpointID)
	}

	s.Aliases[endpointID] = endpointID
	s.Endpoints[endpointID] = eptInfo

	return nil
}

func (s *l2tpEndpointStore) AddEndpointAlias(endpointID, alias string) error {
	if _, ok := s.Aliases[endpointID]; !ok {
		return errors.Errorf("network not found: %s", endpointID)
	}

	s.Aliases[alias] = endpointID

	return nil
}

func (s *l2tpEndpointStore) RemoveEndpoint(endpointID string) error {
	translatedID, ok := s.Aliases[endpointID]
	if !ok {
		return errors.Errorf("endpoint not found: %s", endpointID)
	}

	delete(s.Endpoints, translatedID)

	for alias, target := range s.Aliases {
		if target == translatedID {
			delete(s.Aliases, alias)
		}
	}

	return nil
}

func (s *l2tpEndpointStore) GetEndpoint(netID string) (*l2tpEndpoint, error) {
	translatedID, ok := s.Aliases[netID]
	if !ok {
		return nil, errors.Errorf("endpoint not found: %s", netID)
	}

	if netInfo, ok := s.Endpoints[translatedID]; ok {
		return netInfo, nil
	}

	return nil, errors.Errorf("endpoint not found: %s", netID)
}

func (s *l2tpEndpointStore) GetEndpoints() []*l2tpEndpoint {
	var out []*l2tpEndpoint
	for _, eptInfo := range s.Endpoints {
		out = append(out, eptInfo)
	}

	return out
}

type l2tpEndpoint struct {
	ID           string
	Name         string
	ConnName     string
	PPPOptFile   string
	PPPDevName   string
	AssignedCIDR string
	AssignedIP   string
	NetworkOpts  *L2TPNetworkConfig
}

func NewL2TPEndpoint(netInfo *l2tpNetwork) *l2tpEndpoint {
	return &l2tpEndpoint{
		Name:        fmt.Sprintf("%d_%s", netInfo.Count, netInfo.PoolID),
		NetworkOpts: netInfo.NetworkOpts,
	}
}

func (e *l2tpEndpoint) setup() error {
	e.ConnName = e.Name + "-connection"
	e.PPPDevName = ("ppp" + e.Name)[:15]
	e.PPPOptFile = pppOptsDir + e.Name + "." + ".client"

	return nil
}

func (e *l2tpEndpoint) GetPppConfig() string {
	cfg := ""

	if e.NetworkOpts.PPPIPCPAcceptLocal {
		cfg += "\nipcp-accept-local"
	}
	if e.NetworkOpts.PPPIPCPAcceptRemote {
		cfg += "\nipcp-accept-remote"
	}
	if e.NetworkOpts.PPPRefuseEAP {
		cfg += "\nrefuse-eap"
	}
	if e.NetworkOpts.PPPRequireMSChapV2 {
		cfg += "\nrequire-mschap-v2"
	}
	if e.NetworkOpts.PPPNoccp {
		cfg += "\nnoccp"
	}
	if e.NetworkOpts.PPPNoauth {
		cfg += "\nnoauth"
	}

	cfg += fmt.Sprintf("\nifname %s", e.PPPDevName)
	cfg += fmt.Sprintf("\nname %s", e.NetworkOpts.PPPUsername)
	cfg += fmt.Sprintf("\npassword %s", e.NetworkOpts.PPPPassword)
	cfg += fmt.Sprintf("\nmtu %s", e.NetworkOpts.PPPMTU)
	cfg += fmt.Sprintf("\nmru %s", e.NetworkOpts.PPPMRU)
	cfg += fmt.Sprintf("\nidle %s", e.NetworkOpts.PPPIdle)
	cfg += fmt.Sprintf("\nconnect-delay %s", e.NetworkOpts.PPPConnectDelay)

	if e.NetworkOpts.PPPDebug {
		cfg += "\ndebug"
	}

	if e.NetworkOpts.PPPDefaultRoute {
		cfg += "\ndefaultroute"
	}
	if e.NetworkOpts.PPPUsepeerdns {
		cfg += "\nusepeerdns"
	}
	if e.NetworkOpts.PPPLock {
		cfg += "\nlock"
	}

	return cfg
}

func (e *l2tpEndpoint) GetXl2tpConfig() []string {
	return []string{
		fmt.Sprintf("lns=%s", e.NetworkOpts.LNSAddr),
		fmt.Sprintf("pppoptfile=%s", e.PPPOptFile),
	}
}

func (s *l2tpState) load() (err error) {
	defer func() {
		if err == store.ErrKeyNotFound {
			err = nil
		}
		if err != nil {
			log.G(s.ctx).Error("failed to load l2tp state, erasing key", zap.Error(err))
			delErr := s.storage.Delete("state")
			if delErr != nil {
				log.G(s.ctx).Error("could not cleanup l2tp state", zap.Error(delErr))
			}
		}
	}()

	exists, err := s.storage.Exists("state")
	if err != nil || !exists {
		return
	}

	data, err := s.storage.Get("state")
	if err != nil {
		return
	}

	err = json.Unmarshal(data.Value, s)
	if err != nil {
		return
	}

	for _, net := range s.Networks {
		for _, ept := range net.Store.Endpoints {
			msg := fmt.Sprintf("Starting with network %s, endpoint %s (%s)", net.ID, ept.ID, ept.AssignedIP)
			log.G(s.ctx).Info(msg)
		}
	}

	return
}
