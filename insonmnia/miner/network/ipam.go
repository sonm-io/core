package network

import (
	"context"
	"io/ioutil"
	"os/exec"
	"time"

	"os"

	"github.com/docker/go-plugins-helpers/ipam"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

type IPAMDriver struct {
	ctx     context.Context
	counter int
	store   *networkInfoStore
}

func NewIPAMDriver(ctx context.Context, store *networkInfoStore) *IPAMDriver {
	return &IPAMDriver{
		ctx:   ctx,
		store: store,
	}
}

func (d *IPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	log.G(d.ctx).Info("received RequestPool request", zap.Any("request", request))

	opts, err := parseOptsIPAM(request)
	if err != nil {
		log.G(d.ctx).Error("failed to parse options", zap.Error(err))
		return nil, errors.Wrap(err, "failed to parse options")
	}

	netInfo := newNetworkInfo(opts)
	if err := netInfo.Setup(); err != nil {
		log.G(d.ctx).Error("failed to setup network", zap.Error(err))
		return nil, err
	}

	if err := d.store.AddNetwork(netInfo.PoolID, netInfo); err != nil {
		log.G(d.ctx).Error("failed to add network to store", zap.Error(err))
		return nil, err
	}

	return &ipam.RequestPoolResponse{PoolID: netInfo.PoolID, Pool: opts.Subnet}, nil
}

func (d *IPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	log.G(d.ctx).Info("received RequestAddress request", zap.Any("request", request))

	netInfo, err := d.store.GetNetwork(request.PoolID)
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	eptInfo := newEndpointInfo(netInfo)
	if err := eptInfo.setup(); err != nil {
		log.G(d.ctx).Error("failed to setup endpoint", zap.String("pool_id", netInfo.PoolID),
			zap.String("network_id", netInfo.ID), zap.Error(err))
		return nil, err
	}

	netInfo.ConnInc()

	var (
		pppCfg       = eptInfo.GetPppConfig()
		xl2tpdCfg    = eptInfo.GetXl2tpConfig()
		addCfgCmd    = exec.Command("xl2tpd-control", "add", eptInfo.ConnName, xl2tpdCfg[0], xl2tpdCfg[1])
		setupConnCmd = exec.Command("xl2tpd-control", "connect", eptInfo.ConnName)
	)
	log.G(d.ctx).Info("creating ppp options file", zap.String("network_id", netInfo.ID),
		zap.String("ppo_opt_file", eptInfo.PPPOptFile))
	if err := ioutil.WriteFile(eptInfo.PPPOptFile, []byte(pppCfg), 0644); err != nil {
		log.G(d.ctx).Error("failed to create ppp options file", zap.String("network_id", netInfo.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to create ppp options file for network %s, config is `%s`",
			netInfo.ID, xl2tpdCfg)
	}

	log.G(d.ctx).Info("adding xl2tp connection config", zap.String("network_id", netInfo.ID),
		zap.String("endpoint_name", eptInfo.Name), zap.Any("config", xl2tpdCfg))
	if err := addCfgCmd.Run(); err != nil {
		log.G(d.ctx).Error("failed to add xl2tpd config", zap.String("network_id", netInfo.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd connection config for network %s, config is `%s`",
			netInfo.ID, xl2tpdCfg)
	}

	var (
		linkEvents  = make(chan netlink.LinkUpdate)
		linkStopper = make(chan struct{})
		addrEvents  = make(chan netlink.AddrUpdate)
		addrStopper = make(chan struct{})
	)
	if err := netlink.LinkSubscribe(linkEvents, linkStopper); err != nil {
		log.G(d.ctx).Error("failed to subscribe to netlink", zap.String("network_id", netInfo.ID),
			zap.Error(err))
		return nil, errors.Wrapf(err, "failed to subscribe to netlink: %s", err)
	}

	if err := netlink.AddrSubscribe(addrEvents, addrStopper); err != nil {
		log.G(d.ctx).Error("failed to subscribe to netlink", zap.String("network_id", netInfo.ID),
			zap.String("endpoint_name", eptInfo.Name), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to subscribe to netlink: %s", err)
	}

	log.G(d.ctx).Info("setting up xl2tpd connection", zap.String("connection_name", eptInfo.ConnName),
		zap.String("network_id", netInfo.ID), zap.String("endpoint_name", eptInfo.Name))
	if err := setupConnCmd.Run(); err != nil {
		log.G(d.ctx).Error("xl2tpd failed to setup connection", zap.String("network_id", netInfo.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd config for network %s, config is `%s`",
			netInfo.ID, xl2tpdCfg)
	}

	assignedCIDR, err := d.getAssignedCIDR(eptInfo.PPPDevName, linkEvents, addrEvents, linkStopper, addrStopper)
	if err != nil {
		log.G(d.ctx).Error("failed to get assigned IP", zap.String("network_id", netInfo.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get assigned IP")
	}

	log.G(d.ctx).Info("received IP", zap.String("network_id", netInfo.ID),
		zap.String("ip", assignedCIDR))

	eptInfo.AssignedIP, _ = getAddrFromCIDR(assignedCIDR)
	eptInfo.AssignedCIDR = assignedCIDR

	if err := netInfo.store.AddEndpoint(eptInfo.AssignedIP, eptInfo); err != nil {
		log.G(d.ctx).Error("failed to add endpoint", zap.Error(err))
		return nil, err
	}

	return &ipam.RequestAddressResponse{Address: eptInfo.AssignedCIDR}, nil
}

func (d *IPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	log.G(d.ctx).Info("received ReleasePool request", zap.Any("request", request))

	netInfo, err := d.store.GetNetwork(request.PoolID)
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return errors.Wrap(err, "failed to get network info")
	}

	// Normally there won't be any endpoints: `ReleasePool` is called after releasing all addresses.
	for _, eptInfo := range netInfo.store.GetEndpoints() {
		if err := d.removeEndpoint(netInfo, eptInfo); err != nil {
			log.G(d.ctx).Error("xl2tpd failed to removeEndpoint", zap.String("pool_id", netInfo.PoolID),
				zap.String("network_id", netInfo.ID), zap.Error(err))
		}
	}

	if err := d.store.RemoveNetwork(request.PoolID); err != nil {
		log.G(d.ctx).Error("failed to remove network", zap.String("pool_id", netInfo.PoolID),
			zap.String("network_id", netInfo.ID), zap.Error(err))
		return err
	}

	return nil
}

func (d *IPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	log.G(d.ctx).Info("received ReleaseAddress request", zap.Any("request", request))

	netInfo, err := d.store.GetNetwork(request.PoolID)
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return errors.Wrap(err, "failed to get network")
	}

	eptInfo, err := netInfo.store.GetEndpoint(request.Address)
	if err != nil {
		log.G(d.ctx).Error("failed to get endpoint", zap.String("pool_id", request.PoolID),
			zap.String("network_id", netInfo.ID), zap.String("ip", request.Address), zap.Error(err))
		return errors.Wrap(err, "failed to get endpoint")
	}

	if err := d.removeEndpoint(netInfo, eptInfo); err != nil {
		log.G(d.ctx).Error("xl2tpd failed to removeEndpoint", zap.String("pool_id", netInfo.PoolID),
			zap.String("network_id", netInfo.ID), zap.Error(err))
		return errors.Wrap(err, "xl2tpd failed to removeEndpoint")
	}

	return nil
}

func (d *IPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	log.G(d.ctx).Info("received GetCapabilities request")
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (d *IPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	log.G(d.ctx).Info("received GetDefaultAddressSpaces request")
	return &ipam.AddressSpacesResponse{}, nil
}

func (d *IPAMDriver) getAssignedCIDR(devName string, link chan netlink.LinkUpdate, addr chan netlink.AddrUpdate,
	linkStopper, addrStopper chan struct{}) (string, error) {
	var (
		linkIndex  int
		assignedIP string
		timeout    = time.Second * 10
	)

	linkTicker := time.NewTicker(timeout)
	defer linkTicker.Stop()

	for {
		var done bool
		select {
		case update := <-link:
			if update.Attrs().Name == devName {
				linkIndex = update.Link.Attrs().Index
				done = true
			}
		case <-linkTicker.C:
			return "", errors.New("failed to receive link update: timeout")
		}

		if done {
			break
		}
	}

	linkStopper <- struct{}{}

	addrTicker := time.NewTicker(timeout)
	defer addrTicker.Stop()

	for {
		var done bool
		select {
		case update := <-addr:
			if update.LinkIndex == linkIndex {
				assignedIP = update.LinkAddress.String()
				done = true
			}
		case <-addrTicker.C:
			return "", errors.New("failed to receive addr update: timeout")
		}

		if done {
			break
		}
	}

	addrStopper <- struct{}{}

	return assignedIP, nil
}

func (d *IPAMDriver) removeEndpoint(netInfo *networkInfo, eptInfo *endpointInfo) error {
	disconnectCmd := exec.Command("xl2tpd-control", "disconnect", eptInfo.ConnName)
	if err := disconnectCmd.Run(); err != nil {
		return errors.Wrapf(err, "xl2rpd failed to close connection %s", eptInfo.ConnName)
	}

	if err := os.Remove(eptInfo.PPPOptFile); err != nil {
		return errors.Wrapf(err, "failed to remove ppp opts file %s", eptInfo.PPPOptFile)
	}

	if err := netInfo.store.RemoveEndpoint(eptInfo.AssignedIP); err != nil {
		return errors.Wrap(err, "failed to remove endpoint from store")
	}

	return nil
}
