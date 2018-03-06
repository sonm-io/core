package network

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/docker/go-plugins-helpers/ipam"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type IPAMDriver struct {
	// NOTE: `mu` is shared with L2TPNeworktDriver.
	mu      *sync.Mutex
	counter int
	state   *l2tpState
	logger  *zap.SugaredLogger
}

func NewIPAMDriver(ctx context.Context, mu *sync.Mutex, state *l2tpState) *IPAMDriver {
	return &IPAMDriver{
		mu:     mu,
		state:  state,
		logger: log.S(ctx).With("source", "l2tp/ipam"),
	}
}

func (d *IPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	d.logger.Infow("received RequestPool request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	opts, err := parseOptsIPAM(request)
	if err != nil {
		d.logger.Errorw("failed to parse options", zap.Error(err))
		return nil, errors.Wrap(err, "failed to parse options")
	}

	l2tpNet := newL2tpNetwork(opts)
	if err := l2tpNet.Setup(); err != nil {
		d.logger.Errorw("failed to setup network", zap.Error(err))
		return nil, err
	}

	if err := d.state.AddNetwork(l2tpNet.PoolID, l2tpNet); err != nil {
		d.logger.Errorw("failed to add network to state", zap.Error(err))
		return nil, err
	}

	return &ipam.RequestPoolResponse{PoolID: l2tpNet.PoolID, Pool: opts.Subnet}, nil
}

func (d *IPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	d.logger.Infow("received RequestAddress request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	if len(request.Address) > 0 {
		d.logger.Errorw("requests for specific addresses are not supported")
		return nil, errors.New("requests for specific addresses are not supported")
	}

	l2tpNet, err := d.state.GetNetwork(request.PoolID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	// The first RequestAddress() call gets gateway IP for the network, which is not required
	// for PPP interfaces.
	if l2tpNet.NeedsGateway {
		l2tpNet.NeedsGateway = false
		d.logger.Infow("allocated fake gateway", zap.String("pool_id", request.PoolID))
		return &ipam.RequestAddressResponse{Address: l2tpNet.NetworkOpts.Subnet}, nil
	}

	l2tpEpt := NewL2TPEndpoint(l2tpNet)
	if err := l2tpEpt.setup(); err != nil {
		d.logger.Errorw("failed to setup endpoint", zap.String("pool_id", l2tpNet.PoolID),
			zap.String("network_id", l2tpNet.ID), zap.Error(err))
		return nil, err
	}

	l2tpNet.ConnInc()

	var (
		pppCfg       = l2tpEpt.GetPppConfig()
		xl2tpdCfg    = l2tpEpt.GetXl2tpConfig()
		addCfgCmd    = exec.Command("xl2tpd-control", "add", l2tpEpt.ConnName, xl2tpdCfg[0], xl2tpdCfg[1])
		setupConnCmd = exec.Command("xl2tpd-control", "connect", l2tpEpt.ConnName)
	)
	d.logger.Infow("creating ppp options file", zap.String("network_id", l2tpNet.ID),
		zap.String("ppo_opt_file", l2tpEpt.PPPOptFile))
	if err := ioutil.WriteFile(l2tpEpt.PPPOptFile, []byte(pppCfg), 0644); err != nil {
		d.logger.Errorw("failed to create ppp options file", zap.String("network_id", l2tpNet.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to create ppp options file for network %s, config is `%s`",
			l2tpNet.ID, xl2tpdCfg)
	}

	d.logger.Infow("adding xl2tp connection config", zap.String("network_id", l2tpNet.ID),
		zap.String("endpoint_name", l2tpEpt.Name), zap.Any("config", xl2tpdCfg))
	if err := addCfgCmd.Run(); err != nil {
		d.logger.Errorw("failed to add xl2tpd config", zap.String("network_id", l2tpNet.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd connection config for network %s, config is `%s`",
			l2tpNet.ID, xl2tpdCfg)
	}

	d.logger.Infow("setting up xl2tpd connection", zap.String("connection_name", l2tpEpt.ConnName),
		zap.String("network_id", l2tpNet.ID), zap.String("endpoint_name", l2tpEpt.Name))
	if err := setupConnCmd.Run(); err != nil {
		d.logger.Errorw("xl2tpd failed to setup connection", zap.String("network_id", l2tpNet.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrapf(err, "failed to add xl2tpd config for network %s, config is `%s`",
			l2tpNet.ID, xl2tpdCfg)
	}

	assignedCIDR, err := d.getAssignedCIDR(l2tpEpt.PPPDevName)
	if err != nil {
		d.logger.Errorw("failed to get assigned IP", zap.String("network_id", l2tpNet.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get assigned IP")
	}

	d.logger.Infow("received IP", zap.String("network_id", l2tpNet.ID),
		zap.String("ip", assignedCIDR))

	l2tpEpt.AssignedIP, _ = getAddrFromCIDR(assignedCIDR)
	l2tpEpt.AssignedCIDR = assignedCIDR
	l2tpNet.Endpoint = l2tpEpt

	return &ipam.RequestAddressResponse{Address: l2tpEpt.AssignedCIDR}, nil
}

func (d *IPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	d.logger.Infow("received ReleasePool request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	l2tpNet, err := d.state.GetNetwork(request.PoolID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return errors.Wrap(err, "failed to get network info")
	}

	if l2tpNet.Endpoint == nil {
		d.logger.Errorw("network's endpoint is nil", zap.String("network_id", l2tpNet.ID))
		return nil
	}

	if err := d.removeEndpoint(l2tpNet, l2tpNet.Endpoint); err != nil {
		d.logger.Errorw("xl2tpd failed to remove endpoint", zap.String("pool_id", l2tpNet.PoolID),
			zap.String("network_id", l2tpNet.ID), zap.Error(err))
	}

	if err := d.state.RemoveNetwork(request.PoolID); err != nil {
		d.logger.Errorw("failed to remove network", zap.String("pool_id", l2tpNet.PoolID),
			zap.String("network_id", l2tpNet.ID), zap.Error(err))
		return err
	}

	return nil
}

func (d *IPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	d.logger.Infow("received ReleaseAddress request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}

func (d *IPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	d.logger.Infow("received GetCapabilities request")
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (d *IPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	d.logger.Infow("received GetDefaultAddressSpaces request")
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return &ipam.AddressSpacesResponse{}, nil
}

func (d *IPAMDriver) getAssignedCIDR(devName string) (string, error) {
	time.Sleep(time.Second * 7)
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range ifaces {
		if i.Name == devName {
			addrs, err := i.Addrs()
			if err != nil {
				return "", err
			}

			if len(addrs) < 1 {
				return "", errors.New("no addresses assigned!")
			}

			return addrs[0].String(), nil
		}
	}

	return "", errors.Errorf("device %s not found", devName)
}

func (d *IPAMDriver) removeEndpoint(l2tpNet *l2tpNetwork, l2tpEpt *l2tpEndpoint) error {
	disconnectCmd := exec.Command("xl2tpd-control", "disconnect", l2tpEpt.ConnName)
	if err := disconnectCmd.Run(); err != nil {
		return errors.Wrapf(err, "xl2rpd failed to close connection %s", l2tpEpt.ConnName)
	}

	if err := os.Remove(l2tpEpt.PPPOptFile); err != nil {
		return errors.Wrapf(err, "failed to remove ppp opts file %s", l2tpEpt.PPPOptFile)
	}

	return nil
}
