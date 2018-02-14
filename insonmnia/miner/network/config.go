package network

import (
	"net"

	"crypto/md5"
	"encoding/hex"

	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/jinzhu/configor"
	"github.com/pkg/errors"
)

type config struct {
	LNSAddr         string `required:"true" yaml:"lns_addr"`
	Subnet          string `required:"true" yaml:"subnet"`
	PPPUsername     string `required:"true" yaml:"ppp_username"`
	PPPPassword     string `required:"true" yaml:"ppp_password"`
	PPPMTU          string `required:"true" yaml:"ppp_mtu" default:"1410"`
	PPPMRU          string `required:"true" yaml:"ppp_mru" default:"1410"`
	PPPIdle         string `required:"true" yaml:"ppp_idle" default:"1800"`
	PPPConnectDelay string `required:"true" yaml:"ppp_connect_delay" default:"5000"`

	PPPDebug            bool `required:"true" yaml:"ppp_debug" default:"true"`
	PPPNoauth           bool `required:"true" yaml:"ppp_noauth" default:"true"`
	PPPNoccp            bool `required:"true" yaml:"ppp_noccp" default:"true"`
	PPPDefaultRoute     bool `required:"true" yaml:"ppp_default_route" default:"true"`
	PPPUsepeerdns       bool `required:"true" yaml:"ppp_use_peer_dns" default:"true"`
	PPPLock             bool `required:"true" yaml:"ppp_lock" default:"true"`
	PPPIPCPAcceptLocal  bool `required:"true" yaml:"ppp_ipcp_accept_local" default:"true"`
	PPPIPCPAcceptRemote bool `required:"true" yaml:"ppp_ipcp_accept_remote" default:"true"`
	PPPRefuseEAP        bool `required:"true" yaml:"ppp_refuse_eap" default:"true"`
	PPPRequireMSChapV2  bool `required:"true" yaml:"ppp_require_mschap_v2" default:"true"`
}

func (o *config) GetHash() string {
	hasher := md5.New()
	hasher.Write([]byte(o.LNSAddr + o.PPPUsername + o.PPPPassword))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (o *config) validate() error {
	if ip := net.ParseIP(o.LNSAddr); ip == nil {
		return errors.Errorf("failed to parse lns_addr `%s` to IP", o.LNSAddr)
	}

	if _, _, err := net.ParseCIDR(o.Subnet); err != nil {
		return errors.Wrapf(err, "failed to parse Subnet `%s` to CIDR", o.Subnet)
	}

	if len(o.PPPUsername) < 1 {
		return errors.New("empty PPP username")
	}

	if len(o.PPPPassword) < 1 {
		return errors.New("empty PPP password")
	}

	return nil
}

func parseOptsIPAM(request *ipam.RequestPoolRequest) (*config, error) {
	path, ok := request.Options["config"]
	if !ok {
		return nil, errors.New("config path not provided")
	}

	cfg := &config{}
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseOptsNetwork(request *network.CreateNetworkRequest) (*config, error) {

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

	cfg := &config{}
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
