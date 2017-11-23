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

type EthConfig struct {
	PrivateKey string `required:"true" yaml:"private_key"`
}

type LocatorConfig struct {
	Address string `required:"true" yaml:"address"`
	Period  int    `required:"true" default:"300" yaml:"period"`
}

type StoreConfig struct {
	Type     string `required:"true" default:"boltdb" yaml:"type"`
	Endpoint string `required:"true" default:"/tmp/sonm/boltdb" yaml:"endpoint"`
	Bucket   string `required:"true" default:"sonm" yaml:"bucket"`
	KeyFile  string `yaml:"key_file"`
	CertFile string `yaml:"cert_file"`
}

type ClusterConfig struct {
	Store                        StoreConfig `yaml:"store"`
	Failover                     bool        `yaml:"failover"`
	GrpcEndpoint                 string      `yaml:"grpc_endpoint"`
	LeaderKey                    string      `yaml:"leader_key" default:"sonm/hub/leader"`
	MemberListKey                string      `yaml:"member_list_key" default:"sonm/hub/list"`
	SynchronizableEntitiesPrefix string      `yaml:"sync_prefix" default:"sonm/hub/sync"`
	LeaderTTL                    uint64      `yaml:"leader_ttl" default:"20"`
	AnnouncePeriod               uint64      `yaml:"announce_period" default:"10"`
	AnnounceTTL                  uint64      `yaml:"announce_ttl" default:"20"`
	MemberGCPeriod               uint64      `yaml:"member_gc_period" default:"60"`
}

type HubConfig struct {
	Endpoint      string         `required:"true" yaml:"endpoint"`
	GatewayConfig *GatewayConfig `yaml:"gateway"`
	Bootnodes     []string       `required:"false" yaml:"bootnodes"`
	Logging       LoggingConfig  `yaml:"logging"`
	Eth           EthConfig      `yaml:"ethereum"`
	Locator       LocatorConfig  `yaml:"locator"`
	Cluster       ClusterConfig  `yaml:"cluster"`
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
