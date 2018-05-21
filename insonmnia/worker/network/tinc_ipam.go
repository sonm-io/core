package network

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"

	"github.com/docker/go-plugins-helpers/ipam"

	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type TincIPAMDriver struct {
	*TincNetworkState
	logger *zap.SugaredLogger
}

func NewTincIPAMDriver(ctx context.Context, state *TincNetworkState, config *TincNetworkConfig) (*TincIPAMDriver, error) {
	return &TincIPAMDriver{
		TincNetworkState: state,
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
	t.logger.Infow("received RequestPool request", zap.Any("request", request))

	n, err := t.netByIPAMOptions(request.Options)
	if err != nil {
		return nil, err
	}
	return &ipam.RequestPoolResponse{
		PoolID: n.NodeID,
		Pool:   n.Pool.String(),
		Data:   request.Options,
	}, nil
}

func (t *TincIPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	t.logger.Infow("received ReleasePool request", zap.Any("request", request))
	return nil
}

func (t *TincIPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	t.logger.Infow("received RequestAddress request", zap.Any("request", request))

	n, err := t.netByID(request.PoolID)
	if err != nil {
		return nil, err
	}
	t.logger.Debugw("fetched network", zap.Any("network", n))

	mask, _ := n.Pool.Mask.Size()

	if mask == 0 {
		t.logger.Errorf("invalid subnet specified for pool %s", n.Pool.String())
		return nil, errors.New("invalid subnet")
	}

	ty, ok := request.Options["RequestAddressType"]
	if ok && ty == "com.docker.network.gateway" {
		ip := make(net.IP, len(n.Pool.IP))
		copy(ip, n.Pool.IP)
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

	addrs, err := n.OccupiedIPs(t.ctx)
	t.logger.Debugw("fetched occupied ips", zap.Any("ips", addrs))
	if err != nil {
		return nil, err
	}

	ip, err := getRandomIP(addrs, n.Pool)
	if err != nil {
		return nil, err
	}

	return &ipam.RequestAddressResponse{
		Address: ip.ToCommon().String() + "/" + fmt.Sprint(mask),
	}, nil
}

func (t *TincIPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	t.logger.Infow("received ReleaseAddress request", zap.Any("request", request))
	return nil
}

func getRandomIP(occupied map[IP4]struct{}, ipNet *net.IPNet) (IP4, error) {
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		return IP4{}, errors.New("invalid mask")
	}
	if len(occupied) >= 1<<uint(32-ones) {
		return IP4{}, errors.New("pool is full")
	}
	var ip IP4
	for i := 0; i < 1000; i++ {
		ip = randomIP(ipNet, bits-ones)
		if _, ok := occupied[ip]; !ok {
			return ip, nil
		}
	}
	return ip, errors.New("give up")
}

func randomIP(ipNet *net.IPNet, addrBits int) IP4 {
	var r uint32
	for {
		r = uint32(rand.Int31n(1 << uint(addrBits)))
		if (r&0xff) != 0xff && r != 0 {
			break
		}
	}
	ip := newIP4(ipNet.IP)
	ip.a += byte(r & 0xff000000 >> 24)
	ip.b += byte(r & 0xff0000 >> 16)
	ip.c += byte(r & 0xff00 >> 8)
	ip.d += byte(r & 0xff)
	return ip
}
