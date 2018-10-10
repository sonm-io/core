package network

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/jinzhu/configor"
)

type L2TPConfig struct {
	Enabled        bool   `yaml:"enabled"`
	NetSocketPath  string `required:"false" yaml:"net_socket_path" default:"/run/docker/plugins/l2tp_net.sock"`
	IPAMSocketPath string `required:"false" yaml:"ipam_socket_path" default:"/run/docker/plugins/l2tp_ipam.sock"`
	ConfigDir      string `required:"false" yaml:"config_dir" default:"/tmp/sonm/l2tp/"`
	StatePath      string `required:"false" yaml:"state_path" default:"/tmp/sonm/l2tp/l2tp_network_state"`
}

type l2tpNetworkConfig struct {
	ID                  string `required:"true" yaml:"id"`
	LNSAddr             string `required:"true" yaml:"lns_addr"`
	Subnet              string `required:"true" yaml:"subnet"`
	PPPUsername         string `required:"false" yaml:"ppp_username"`
	PPPPassword         string `required:"false" yaml:"ppp_password"`
	PPPMTU              string `required:"false" yaml:"ppp_mtu" default:"1410"`
	PPPMRU              string `required:"false" yaml:"ppp_mru" default:"1410"`
	PPPIdle             string `required:"false" yaml:"ppp_idle" default:"1800"`
	PPPConnectDelay     string `required:"false" yaml:"ppp_connect_delay" default:"5000"`
	PPPDebug            bool   `required:"false" yaml:"ppp_debug" default:"true"`
	PPPNoauth           bool   `required:"false" yaml:"ppp_noauth" default:"true"`
	PPPNoccp            bool   `required:"false" yaml:"ppp_noccp" default:"true"`
	PPPDefaultRoute     bool   `required:"false" yaml:"ppp_default_route" default:"true"`
	PPPUsepeerdns       bool   `required:"false" yaml:"ppp_use_peer_dns" default:"true"`
	PPPLock             bool   `required:"false" yaml:"ppp_lock" default:"true"`
	PPPIPCPAcceptLocal  bool   `required:"false" yaml:"ppp_ipcp_accept_local" default:"true"`
	PPPIPCPAcceptRemote bool   `required:"false" yaml:"ppp_ipcp_accept_remote" default:"true"`
	PPPRefuseEAP        bool   `required:"false" yaml:"ppp_refuse_eap" default:"true"`
	PPPRequireMSChapV2  bool   `required:"false" yaml:"ppp_require_mschap_v2" default:"true"`
}

func (o *l2tpNetworkConfig) PoolID() string {
	hasher := md5.New()
	hasher.Write([]byte(o.ID + o.LNSAddr))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (o *l2tpNetworkConfig) validate() error {
	if ip := net.ParseIP(o.LNSAddr); ip == nil {
		return fmt.Errorf("failed to parse lns_addr `%s` to IP", o.LNSAddr)
	}

	if _, _, err := net.ParseCIDR(o.Subnet); err != nil {
		return fmt.Errorf("failed to parse Subnet `%s` to CIDR: %v", o.Subnet, err)
	}

	return nil
}

func parseOptsIPAM(request *ipam.RequestPoolRequest) (*l2tpNetworkConfig, error) {
	path, ok := request.Options["config"]
	if !ok {
		return nil, errors.New("config path not provided")
	}

	cfg := &l2tpNetworkConfig{}
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseOptsNetwork(request *network.CreateNetworkRequest) (*l2tpNetworkConfig, error) {
	rawOpts, ok := request.Options["com.docker.network.generic"]
	if !ok {
		return nil, errors.New("no options provided")
	}

	opts, ok := rawOpts.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid options")
	}

	path, ok := opts["config"]
	if !ok {
		return nil, errors.New("config path not provided")
	}

	cfg := &l2tpNetworkConfig{}
	if err := configor.Load(cfg, path.(string)); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getAddrFromCIDR(cidr string) (string, error) {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}
