package blockchain

import (
	"net/url"

	"github.com/ethereum/go-ethereum/common"
)

// Config represents SONM blockchain configuration structure that can act as a
// building block for more complex configs.
type Config struct {
	Endpoint             url.URL
	SidechainEndpoint    url.URL
	ContractRegistryAddr common.Address
}

func (m *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg struct {
		MasterchainEndpoint  string         `yaml:"endpoint"`
		SidechainEndpoint    string         `yaml:"sidechain_endpoint"`
		ContractRegistryAddr common.Address `yaml:"contract_registry"`
	}

	if err := unmarshal(&cfg); err != nil {
		return err
	}

	if len(cfg.MasterchainEndpoint) == 0 {
		cfg.MasterchainEndpoint = defaultMasterchainEndpoint
	}

	if len(cfg.SidechainEndpoint) == 0 {
		cfg.SidechainEndpoint = defaultSidechainEndpoint
	}

	endpoint, err := url.Parse(cfg.MasterchainEndpoint)
	if err != nil {
		return err
	}

	sidechainEndpoint, err := url.Parse(cfg.SidechainEndpoint)
	if err != nil {
		return err
	}

	m.Endpoint = *endpoint
	m.SidechainEndpoint = *sidechainEndpoint
	m.ContractRegistryAddr = cfg.ContractRegistryAddr

	return nil
}
