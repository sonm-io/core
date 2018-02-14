package network

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type networkInfoStore struct {
	mu       sync.Mutex
	aliases  map[string]string
	networks map[string]*networkInfo
}

func NewNetworkInterface() *networkInfoStore {
	return &networkInfoStore{
		aliases:  make(map[string]string),
		networks: make(map[string]*networkInfo),
	}
}

func (s *networkInfoStore) AddNetwork(netID string, netInfo *networkInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.aliases[netID]; ok {
		return errors.Errorf("network already exists: %s", netID)
	}

	s.aliases[netID] = netID
	s.networks[netID] = netInfo

	return nil
}

func (s *networkInfoStore) AddNetworkAlias(netID, alias string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.aliases[netID]; !ok {
		return errors.Errorf("network not found: %s", netID)
	}

	s.aliases[alias] = netID

	return nil
}

func (s *networkInfoStore) RemoveNetwork(netID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	translatedID, ok := s.aliases[netID]
	if !ok {
		return errors.Errorf("network not found: %s", netID)
	}

	delete(s.networks, translatedID)

	for alias, target := range s.aliases {
		if target == translatedID {
			delete(s.aliases, alias)
		}
	}

	return nil
}

func (s *networkInfoStore) GetNetwork(netID string) (*networkInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	translatedID, ok := s.aliases[netID]
	if !ok {
		return nil, errors.Errorf("network not found: %s", netID)
	}

	if netInfo, ok := s.networks[translatedID]; ok {
		return netInfo, nil
	}

	return nil, errors.Errorf("network not found: %s", netID)
}

func (s *networkInfoStore) GetNetworks() []*networkInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []*networkInfo
	for _, netInfo := range s.networks {
		out = append(out, netInfo)
	}

	return out
}

type networkInfo struct {
	ID          string
	PoolID      string
	count       int
	networkOpts *config
	store       *endpointInfoStore
}

func newNetworkInfo(opts *config) *networkInfo {
	return &networkInfo{
		networkOpts: opts,
		store:       newEndpointStore(),
	}
}

func (i *networkInfo) Setup() error {
	i.PoolID = i.networkOpts.GetHash()

	return nil
}

func (i *networkInfo) ConnInc() {
	i.count++
}

type endpointInfoStore struct {
	mu        sync.Mutex
	aliases   map[string]string
	endpoints map[string]*endpointInfo
}

func newEndpointStore() *endpointInfoStore {
	return &endpointInfoStore{
		aliases:   make(map[string]string),
		endpoints: make(map[string]*endpointInfo),
	}
}

func (s *endpointInfoStore) AddEndpoint(endpointID string, eptInfo *endpointInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.aliases[endpointID]; ok {
		return errors.Errorf("endpoint already exists: %s", endpointID)
	}

	s.aliases[endpointID] = endpointID
	s.endpoints[endpointID] = eptInfo

	return nil
}

func (s *endpointInfoStore) AddEndpointAlias(endpointID, alias string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.aliases[endpointID]; !ok {
		return errors.Errorf("network not found: %s", endpointID)
	}

	s.aliases[alias] = endpointID

	return nil
}

func (s *endpointInfoStore) RemoveEndpoint(endpointID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	translatedID, ok := s.aliases[endpointID]
	if !ok {
		return errors.Errorf("endpoint not found: %s", endpointID)
	}

	delete(s.endpoints, translatedID)

	for alias, target := range s.aliases {
		if target == translatedID {
			delete(s.aliases, alias)
		}
	}

	return nil
}

func (s *endpointInfoStore) GetEndpoint(netID string) (*endpointInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	translatedID, ok := s.aliases[netID]
	if !ok {
		return nil, errors.Errorf("endpoint not found: %s", netID)
	}

	if netInfo, ok := s.endpoints[translatedID]; ok {
		return netInfo, nil
	}

	return nil, errors.Errorf("endpoint not found: %s", netID)
}

func (s *endpointInfoStore) GetEndpoints() []*endpointInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []*endpointInfo
	for _, eptInfo := range s.endpoints {
		out = append(out, eptInfo)
	}

	return out
}

type endpointInfo struct {
	ID           string
	Name         string
	ConnName     string
	PPPOptFile   string
	PPPDevName   string
	AssignedCIDR string
	AssignedIP   string
	networkOpts  *config
}

func newEndpointInfo(netInfo *networkInfo) *endpointInfo {
	return &endpointInfo{
		Name:        fmt.Sprintf("%d_%s", netInfo.count, netInfo.PoolID),
		networkOpts: netInfo.networkOpts,
	}
}

func (i *endpointInfo) setup() error {
	i.ConnName = i.Name + "-connection"
	i.PPPDevName = ("ppp" + i.Name)[:15]
	i.PPPOptFile = pppOptsDir + i.Name + "." + ".client"

	return nil
}

func (i *endpointInfo) GetPppConfig() string {
	cfg := ""

	if i.networkOpts.PPPIPCPAcceptLocal {
		cfg += "\nipcp-accept-local"
	}
	if i.networkOpts.PPPIPCPAcceptRemote {
		cfg += "\nipcp-accept-remote"
	}
	if i.networkOpts.PPPRefuseEAP {
		cfg += "\nrefuse-eap"
	}
	if i.networkOpts.PPPRequireMSChapV2 {
		cfg += "\nrequire-mschap-v2"
	}
	if i.networkOpts.PPPNoccp {
		cfg += "\nnoccp"
	}
	if i.networkOpts.PPPNoauth {
		cfg += "\nnoauth"
	}

	cfg += fmt.Sprintf("\nifname %s", i.PPPDevName)
	cfg += fmt.Sprintf("\nname %s", i.networkOpts.PPPUsername)
	cfg += fmt.Sprintf("\npassword %s", i.networkOpts.PPPPassword)
	cfg += fmt.Sprintf("\nmtu %s", i.networkOpts.PPPMTU)
	cfg += fmt.Sprintf("\nmru %s", i.networkOpts.PPPMRU)
	cfg += fmt.Sprintf("\nidle %s", i.networkOpts.PPPIdle)
	cfg += fmt.Sprintf("\nconnect-delay %s", i.networkOpts.PPPConnectDelay)

	if i.networkOpts.PPPDebug {
		cfg += "\ndebug"
	}

	if i.networkOpts.PPPDefaultRoute {
		cfg += "\ndefaultroute"
	}
	if i.networkOpts.PPPUsepeerdns {
		cfg += "\nusepeerdns"
	}
	if i.networkOpts.PPPLock {
		cfg += "\nlock"
	}

	return cfg
}

func (i *endpointInfo) GetXl2tpConfig() []string {
	return []string{
		fmt.Sprintf("lns=%s", i.networkOpts.LNSAddr),
		fmt.Sprintf("pppoptfile=%s", i.PPPOptFile),
	}
}
