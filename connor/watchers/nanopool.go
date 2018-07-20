package watchers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
)

type ReportedHashrate struct {
	Status bool     `json:"status"`
	Data   []RHData `json:"data"`
}

type RHData struct {
	Worker   string  `json:"worker"`
	Hashrate float64 `json:"hashrate"`
}

type nanopoolWatcher struct {
	mu                 sync.Mutex
	url                string
	addr               []string
	data               map[string]*ReportedHashrate
	hashrateMultiplier float64
}

func NewPoolWatcher(url string, addr []string, hashrateMultiplier float64) PoolWatcher {
	return &nanopoolWatcher{
		url:                url,
		addr:               addr,
		data:               make(map[string]*ReportedHashrate),
		hashrateMultiplier: hashrateMultiplier,
	}
}

func (p *nanopoolWatcher) GetData(addr string) (*ReportedHashrate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	d, ok := p.data[addr]
	if !ok {
		return nil, errors.New("no pool with given addr")
	}

	return d, nil
}

func (p *nanopoolWatcher) Update(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, addr := range p.addr {
		forPool, err := p.getPoolData(addr, p.url)
		if err != nil {
			return err
		}
		p.data[addr] = forPool
	}
	return nil
}

func (p *nanopoolWatcher) getPoolData(addr string, url string) (*ReportedHashrate, error) {
	body, err := fetchBody(url + addr)
	if err != nil {
		return nil, err
	}

	forPool := &ReportedHashrate{}
	err = json.Unmarshal(body, forPool)
	if err != nil {
		return nil, err
	}
	for idx := range forPool.Data {
		forPool.Data[idx].Hashrate *= p.hashrateMultiplier
	}
	return forPool, nil
}
