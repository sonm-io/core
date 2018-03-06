package network

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/libkv/store"
	"github.com/docker/libnetwork/datastore"
	i "github.com/docker/libnetwork/ipam"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"go.uber.org/zap"
)

type TincIPAMDriver struct {
	*TincNetworkState
	allocator *i.Allocator
	logger    *zap.SugaredLogger
}

func NewTincIPAMDriver(ctx context.Context, state *TincNetworkState, config *TincNetworkConfig) (*TincIPAMDriver, error) {
	s, err := datastore.NewDataStore(datastore.LocalScope, &datastore.ScopeCfg{
		Client: datastore.ScopeClientCfg{
			Provider: string(store.BOLTDB),
			Address:  config.StatePath,
			Config: &store.Config{
				Bucket: "sonm_tinc_ipam",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	allocator, err := i.NewAllocator(s, s)
	if err != nil {
		return nil, err
	}

	return &TincIPAMDriver{
		TincNetworkState: state,
		allocator:        allocator,
		logger:           log.S(ctx).With("source", "tinc/ipam"),
	}, nil
}

func (t *TincIPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	t.logger.Info("received GetCapabilities request")
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (t *TincIPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	t.logger.Info("received GetDefaultAddressSpaces request")
	return nil, nil
}

func (t *TincIPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {

	t.logger.Info("received RequestPool request", zap.Any("request", *request))
	t.mu.Lock()
	defer t.mu.Unlock()
	id := uuid.New()
	_, n, err := net.ParseCIDR(request.Pool)
	if err != nil {
		t.logger.Errorf("invalid pool CIDR specified - %s", err)
		return nil, err
	}
	t.Pools[id] = n
	t.sync()
	return &ipam.RequestPoolResponse{
		PoolID: id,
		Pool:   request.Pool,
		Data:   request.Options,
	}, nil
}

func (t *TincIPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	t.logger.Info("received ReleasePool request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.Pools, request.PoolID)
	t.sync()
	return nil
}

func (t *TincIPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	t.logger.Info("received RequestAddress request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()

	pool, ok := t.Pools[request.PoolID]
	if !ok {
		t.logger.Errorf("pool %s not found", request.PoolID)
		return nil, errors.New("pool not found")
	}

	mask, _ := pool.Mask.Size()

	ty, ok := request.Options["RequestAddressType"]
	if ok && ty == "com.docker.network.gateway" {
		ip := make(net.IP, len(pool.IP))
		copy(ip, pool.IP)
		if len(ip) == 4 {
			ip[3]++
		} else {
			ip[15]++
		}
		addr := ip.String() + "/" + fmt.Sprint(mask)
		t.logger.Infof("providing gateway address %s", addr)
		return &ipam.RequestAddressResponse{
			Address: addr,
		}, nil
	}

	if mask == 0 {
		t.logger.Errorf("invalid subnet specified for pool %s", pool.String())
		return nil, errors.New("invalid subnet")
	}
	return &ipam.RequestAddressResponse{
		Address: request.Address + "/" + fmt.Sprint(mask),
	}, nil
}

func (t *TincIPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	t.logger.Info("received ReleaseAddress request", zap.Any("request", request))
	return nil
}
