package network

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/docker/go-plugins-helpers/ipam"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type IPAMDriver struct {
	*l2tpState
	counter int
	logger  *zap.SugaredLogger
}

func NewIPAMDriver(ctx context.Context, state *l2tpState) *IPAMDriver {
	return &IPAMDriver{
		l2tpState: state,
		logger:    log.S(ctx).With("source", "l2tp/ipam"),
	}
}

func (d *IPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	d.logger.Infow("received RequestPool request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.sync()

	opts, err := parseOptsIPAM(request)
	if err != nil {
		d.logger.Errorw("failed to parse options", zap.Error(err))
		return nil, fmt.Errorf("failed to parse options: %v", err)
	}

	n := newL2tpNetwork(opts)
	if err := n.Setup(); err != nil {
		d.logger.Errorw("failed to setup network", zap.Error(err))
		return nil, err
	}

	if err := d.AddNetwork(n.PoolID, n); err != nil {
		d.logger.Errorw("failed to add network to state", zap.Error(err))
		return nil, err
	}

	return &ipam.RequestPoolResponse{PoolID: n.PoolID, Pool: opts.Subnet}, nil
}

func (d *IPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	d.logger.Infow("received RequestAddress request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.sync()

	if len(request.Address) > 0 {
		d.logger.Errorw("requests for specific addresses are not supported")
		return nil, errors.New("requests for specific addresses are not supported")
	}

	n, err := d.GetNetwork(request.PoolID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %v", err)
	}

	// The first RequestAddress() call gets gateway IP for the network, which is not required
	// for PPP interfaces.
	if n.NeedsGateway {
		n.NeedsGateway = false
		d.logger.Infow("allocated fake gateway", zap.String("pool_id", request.PoolID))
		return &ipam.RequestAddressResponse{Address: n.NetworkOpts.Subnet}, nil
	}

	ept := NewL2TPEndpoint(n)
	if err := ept.setup(); err != nil {
		d.logger.Errorw("failed to setup endpoint", zap.String("pool_id", n.PoolID),
			zap.String("network_id", n.ID), zap.Error(err))
		return nil, err
	}

	n.ConnInc()

	var (
		pppCfg       = ept.GetPppConfig()
		xl2tpdCfg    = ept.GetXl2tpConfig()
		addCfgCmd    = exec.Command("xl2tpd-control", "add", ept.ConnName, xl2tpdCfg[0], xl2tpdCfg[1])
		setupConnCmd = exec.Command("xl2tpd-control", "connect", ept.ConnName)
	)
	d.logger.Infow("creating ppp options file", zap.String("network_id", n.ID),
		zap.String("ppo_opt_file", ept.PPPOptFile))
	if err := ioutil.WriteFile(ept.PPPOptFile, []byte(pppCfg), 0644); err != nil {
		d.logger.Errorw("failed to create ppp options file", zap.String("network_id", n.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, fmt.Errorf("failed to create ppp options file for network %s, config is `%s`: %v",
			n.ID, xl2tpdCfg, err)
	}

	d.logger.Infow("adding xl2tp connection config", zap.String("network_id", n.ID),
		zap.String("endpoint_name", ept.Name), zap.Any("config", xl2tpdCfg))
	if err := addCfgCmd.Run(); err != nil {
		d.logger.Errorw("failed to add xl2tpd config", zap.String("network_id", n.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, fmt.Errorf("failed to add xl2tpd connection config for network %s, config is `%s`: %v",
			n.ID, xl2tpdCfg, err)
	}

	d.logger.Infow("setting up xl2tpd connection", zap.String("connection_name", ept.ConnName),
		zap.String("network_id", n.ID), zap.String("endpoint_name", ept.Name))
	if err := setupConnCmd.Run(); err != nil {
		d.logger.Errorw("xl2tpd failed to setup connection", zap.String("network_id", n.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, fmt.Errorf("failed to add xl2tpd config for network %s, config is `%s`: %v",
			n.ID, xl2tpdCfg, err)
	}

	assignedCIDR, err := d.getAssignedCIDR(ept.PPPDevName)
	if err != nil {
		d.logger.Errorw("failed to get assigned IP", zap.String("network_id", n.ID),
			zap.Any("config", xl2tpdCfg), zap.Error(err))
		return nil, fmt.Errorf("failed to get assigned IP: %v", err)
	}

	d.logger.Infow("received IP", zap.String("network_id", n.ID),
		zap.String("ip", assignedCIDR))

	ept.AssignedIP, _ = getAddrFromCIDR(assignedCIDR)
	ept.AssignedCIDR = assignedCIDR
	n.Endpoint = ept

	return &ipam.RequestAddressResponse{Address: ept.AssignedCIDR}, nil
}

func (d *IPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	d.logger.Infow("received ReleasePool request", zap.Any("request", request))
	n, err := d.GetNetwork(request.PoolID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.PoolID), zap.Error(err))
		return fmt.Errorf("failed to get network info: %v", err)
	}

	if n.Endpoint == nil {
		d.logger.Errorw("network's endpoint is nil", zap.String("network_id", n.ID))
		return nil
	}

	if err := d.removeEndpoint(n, n.Endpoint); err != nil {
		d.logger.Errorw("xl2tpd failed to remove endpoint", zap.String("pool_id", n.PoolID),
			zap.String("network_id", n.ID), zap.Error(err))
	}

	if err := d.RemoveNetwork(request.PoolID); err != nil {
		d.logger.Errorw("failed to remove network", zap.String("pool_id", n.PoolID),
			zap.String("network_id", n.ID), zap.Error(err))
		return err
	}

	return nil
}

func (d *IPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	d.logger.Infow("received ReleaseAddress request", zap.Any("request", request))
	return nil
}

func (d *IPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	d.logger.Infow("received GetCapabilities request")
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (d *IPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	d.logger.Infow("received GetDefaultAddressSpaces request")
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
				return "", errors.New("no addresses assigned")
			}

			return addrs[0].String(), nil
		}
	}

	return "", fmt.Errorf("device %s not found", devName)
}

func (d *IPAMDriver) removeEndpoint(n *l2tpNetwork, ept *l2tpEndpoint) error {
	disconnectCmd := exec.Command("xl2tpd-control", "disconnect", ept.ConnName)
	if err := disconnectCmd.Run(); err != nil {
		return fmt.Errorf("xl2rpd failed to close connection %s: %v", ept.ConnName, err)
	}

	if err := os.Remove(ept.PPPOptFile); err != nil {
		return fmt.Errorf("failed to remove ppp opts file %s: %v", ept.PPPOptFile, err)
	}

	return nil
}
