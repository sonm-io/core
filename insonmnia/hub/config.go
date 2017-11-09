package hub

import (
	"github.com/jinzhu/configor"
)

type LoggingConfig struct {
	Level int `required:"true" default:"1"`
}

type GatewayConfig struct {
	Ports []uint16 `required:"true" yaml:"ports"`
}

type MonitoringConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type EthConfig struct {
	PrivateKey string `required:"true" yaml:"private_key"`
}

type LocatorConfig struct {
	Address string `required:"true" yaml:"address"`
	Period  int    `required:"true" default:"300" yaml:"period"`
}

type ClusterConfig struct {
	StoreType     string `required:"true" default:"boltdb"`
	StoreEndpoint string `required:"true" default:"/tmp/sonm/boltdb"`
	StoreBucket   string `required:"true" default:"sonm"`
	Failover      bool   `required:"true" default:"false"`
	GrpcIp        string `required:"false"`
	GrpcPort      int    `required:"true" default:"10002"`
}

type HubConfig struct {
	// TODO: Deprecated - use ClusterConfig's GrpcIp+GrpcPort
	Endpoint      string           `required:"true" yaml:"endpoint"`
	GatewayConfig *GatewayConfig   `yaml:"gateway"`
	Bootnodes     []string         `required:"false" yaml:"bootnodes"`
	Monitoring    MonitoringConfig `required:"true" yaml:"monitoring"`
	Logging       LoggingConfig    `yaml:"logging"`
	Eth           EthConfig        `yaml:"ethereum"`
	Locator       LocatorConfig    `yaml:"locator"`
	Cluster       ClusterConfig    `yaml:"store"`
}

// NewConfig loads a hub config from the specified YAML file.
func NewConfig(path string) (*HubConfig, error) {
	conf := &HubConfig{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// TODO: Currently stubbed for simplifying testing.
type Config interface {
	Endpoint() string
	// Gateway returns optional gateway config.
	Gateway() *GatewayConfig
	MonitoringEndpoint() string
	Logging() LoggingConfig
	Eth() EthConfig
	Locator() LocatorConfig
}
