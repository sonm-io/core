package node

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp"
)

// Config is LocalNode config
type Config interface {
	// HttpBindPort is port to listen for client connection via REST at localhost
	HttpBindPort() uint16
	// BindPort is port to listen for client connection via GRPC at localhost
	BindPort() uint16
	NPPConfig() *npp.Config
	// MarketEndpoint is Marketplace gRPC endpoint
	MarketEndpoint() string
	// HubEndpoint is Hub's gRPC endpoint (not required)
	HubEndpoint() string
	// MetricsListenAddr returns the address that can be used by Prometheus to get
	// metrics.
	MetricsListenAddr() string
	// KeyStorager included into config because of
	// Node instance must know how to open the keystore
	accounts.KeyStorager
	logging.Leveler
}

type nodeConfig struct {
	HttpBindPort uint16 `yaml:"http_bind_port" default:"15031"`
	BindPort     uint16 `yaml:"bind_port" default:"15030"`
}

type marketConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type hubConfig struct {
	Endpoint string `required:"false" yaml:"endpoint"`
}

type logConfig struct {
	Level logging.Level `yaml:"level" required:"true" default:"debug"`
}

type yamlConfig struct {
	Node                    nodeConfig         `yaml:"node"`
	NPPCfg                  npp.Config         `yaml:"npp"`
	Market                  marketConfig       `required:"true" yaml:"market"`
	Log                     logConfig          `required:"true" yaml:"log"`
	Eth                     accounts.EthConfig `required:"false" yaml:"ethereum"`
	Hub                     *hubConfig         `required:"false" yaml:"hub"`
	MetricsListenAddrConfig string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14003"`
}

func (y *yamlConfig) HttpBindPort() uint16 {
	return y.Node.HttpBindPort
}

func (y *yamlConfig) BindPort() uint16 {
	return y.Node.BindPort
}

func (y *yamlConfig) NPPConfig() *npp.Config {
	return &y.NPPCfg
}

func (y *yamlConfig) MarketEndpoint() string {
	return y.Market.Endpoint
}

func (y *yamlConfig) HubEndpoint() string {
	if y.Hub != nil {
		return y.Hub.Endpoint
	}
	return ""
}

func (y *yamlConfig) LogLevel() logging.Level {
	return y.Log.Level
}

func (y *yamlConfig) KeyStore() string {
	return y.Eth.Keystore
}

func (y *yamlConfig) PassPhrase() string {
	return y.Eth.Passphrase
}

func (y *yamlConfig) MetricsListenAddr() string {
	return y.MetricsListenAddrConfig
}

// NewConfig loads localNode config from given .yaml file
func NewConfig(path string) (Config, error) {
	cfg := &yamlConfig{}

	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
