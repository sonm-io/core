package node

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"go.uber.org/zap/zapcore"
)

// Config is LocalNode config
type Config interface {
	// ListenAddress is gRPC endpoint that Node binds to
	ListenAddress() string
	// MarketEndpoint is Marketplace gRPC endpoint
	MarketEndpoint() string
	// HubEndpoint is Hub's gRPC endpoint (not required)
	HubEndpoint() string
	// LocatorEndpoint is Locator service gRPC endpoint
	LocatorEndpoint() string
	// MetricsListenAddr returns the address that can be used by Prometheus to get
	// metrics.
	MetricsListenAddr() string
	// KeyStorager included into config because of
	// Node instance must know how to open the keystore
	accounts.KeyStorager
	logging.Leveler
}

type nodeConfig struct {
	ListenAddr string `required:"true" yaml:"listen_addr"`
}

type marketConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type hubConfig struct {
	Endpoint string `required:"false" yaml:"endpoint"`
}

type logConfig struct {
	Level string `required:"true" default:"debug" yaml:"level"`
}

type locatorConfig struct {
	Endpoint string `required:"true" default:"" yaml:"endpoint"`
}

type yamlConfig struct {
	Node                    nodeConfig         `required:"true" yaml:"node"`
	Market                  marketConfig       `required:"true" yaml:"market"`
	Log                     logConfig          `required:"true" yaml:"log"`
	Locator                 locatorConfig      `required:"true" yaml:"locator"`
	Eth                     accounts.EthConfig `required:"false" yaml:"ethereum"`
	Hub                     *hubConfig         `required:"false" yaml:"hub"`
	MetricsListenAddrConfig string             `yaml:"metrics_listen_addr" default:"127.0.0.1:14003"`
}

func (y *yamlConfig) ListenAddress() string {
	return y.Node.ListenAddr
}

func (y *yamlConfig) MarketEndpoint() string {
	return y.Market.Endpoint
}

func (y *yamlConfig) LocatorEndpoint() string {
	return y.Locator.Endpoint
}

func (y *yamlConfig) HubEndpoint() string {
	if y.Hub != nil {
		return y.Hub.Endpoint
	}
	return ""
}

func (y *yamlConfig) LogLevel() (zapcore.Level, error) {
	return logging.ParseLogLevel(y.Log.Level)
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
