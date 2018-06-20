package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

const (
	pppOptsDir = "/etc/ppp/"
)

type l2tpState struct {
	mu       sync.Mutex
	Aliases  map[string]string
	Networks map[string]*l2tpNetwork
	storage  store.Store
	logger   *zap.SugaredLogger
}

func newL2TPNetworkState(ctx context.Context, path string) (*l2tpState, error) {
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
		logger:   log.S(ctx).With("source", "l2tp/store"),
		storage:  storage,
	}

	if err := state.load(); err != nil {
		return nil, err
	}

	return state, nil
}

func (s *l2tpState) sync() (err error) {
	defer func() {
		if err != nil {
			s.logger.Error("failed to sync l2tp state", zap.Error(err))
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
		return fmt.Errorf("network already exists: %s", netID)
	}

	s.Aliases[netID] = netID
	s.Networks[netID] = netInfo

	return nil
}

func (s *l2tpState) AddNetworkAlias(netID, alias string) error {
	if _, ok := s.Aliases[netID]; !ok {
		return fmt.Errorf("network not found: %s", netID)
	}

	s.Aliases[alias] = netID

	return nil
}

func (s *l2tpState) RemoveNetwork(netID string) error {
	translatedID, ok := s.Aliases[netID]
	if !ok {
		return fmt.Errorf("network not found: %s", netID)
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
		return nil, fmt.Errorf("network not found: %s", netID)
	}

	if netInfo, ok := s.Networks[translatedID]; ok {
		return netInfo, nil
	}

	return nil, fmt.Errorf("network not found: %s", netID)
}

type l2tpNetwork struct {
	ID           string
	PoolID       string
	Count        int
	NetworkOpts  *l2tpNetworkConfig
	Endpoint     *l2tpEndpoint
	NeedsGateway bool
}

func newL2tpNetwork(opts *l2tpNetworkConfig) *l2tpNetwork {
	return &l2tpNetwork{
		NetworkOpts:  opts,
		NeedsGateway: true,
	}
}

func (n *l2tpNetwork) Setup() error {
	n.PoolID = n.NetworkOpts.PoolID()

	return nil
}

func (n *l2tpNetwork) ConnInc() {
	n.Count++
}

type l2tpEndpoint struct {
	ID           string
	Name         string
	ConnName     string
	PPPOptFile   string
	PPPDevName   string
	AssignedCIDR string
	AssignedIP   string
	NetworkOpts  *l2tpNetworkConfig
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
	cfg := new(bytes.Buffer)

	if e.NetworkOpts.PPPIPCPAcceptLocal {
		fmt.Fprint(cfg, "\nipcp-accept-local")
	}
	if e.NetworkOpts.PPPIPCPAcceptRemote {
		fmt.Fprint(cfg, "\nipcp-accept-remote")
	}
	if e.NetworkOpts.PPPRefuseEAP {
		fmt.Fprint(cfg, "\nrefuse-eap")
	}
	if e.NetworkOpts.PPPRequireMSChapV2 {
		fmt.Fprint(cfg, "\nrequire-mschap-v2")
	}
	if e.NetworkOpts.PPPNoccp {
		fmt.Fprint(cfg, "\nnoccp")
	}
	if e.NetworkOpts.PPPNoauth {
		fmt.Fprint(cfg, "\nnoauth")
	}

	fmt.Fprintf(cfg, "\nifname %s", e.PPPDevName)
	fmt.Fprintf(cfg, "\nname %s", e.NetworkOpts.PPPUsername)
	fmt.Fprintf(cfg, "\npassword %s", e.NetworkOpts.PPPPassword)
	fmt.Fprintf(cfg, "\nmtu %s", e.NetworkOpts.PPPMTU)
	fmt.Fprintf(cfg, "\nmru %s", e.NetworkOpts.PPPMRU)
	fmt.Fprintf(cfg, "\nidle %s", e.NetworkOpts.PPPIdle)
	fmt.Fprintf(cfg, "\nconnect-delay %s", e.NetworkOpts.PPPConnectDelay)

	if e.NetworkOpts.PPPDebug {
		fmt.Fprint(cfg, "\ndebug")
	}

	if e.NetworkOpts.PPPDefaultRoute {
		fmt.Fprint(cfg, "\ndefaultroute")
	}
	if e.NetworkOpts.PPPUsepeerdns {
		fmt.Fprint(cfg, "\nusepeerdns")
	}
	if e.NetworkOpts.PPPLock {
		fmt.Fprint(cfg, "\nlock")
	}

	return cfg.String()
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
			s.logger.Error("failed to load l2tp state, erasing key", zap.Error(err))
			delErr := s.storage.Delete("state")
			if delErr != nil {
				s.logger.Error("could not cleanup l2tp state", zap.Error(delErr))
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

	for netID, l2tpNet := range s.Networks {
		if l2tpNet.Endpoint == nil {
			s.logger.Warnw("found a network without endpoint, removing", zap.String("network_id", netID))
			delete(s.Networks, netID)
			continue
		}
		msg := fmt.Sprintf("starting with network %s, endpoint %s (%s)",
			l2tpNet.ID, l2tpNet.Endpoint.ID, l2tpNet.Endpoint.AssignedIP)
		s.logger.Info(msg)
	}

	return
}
