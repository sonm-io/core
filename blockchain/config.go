package blockchain

import (
	"math/big"
	"net/url"

	"github.com/ethereum/go-ethereum/common"
)

// Config represents SONM blockchain configuration structure that can act as a
// building block for more complex configs.
type Config struct {
	Endpoint             url.URL
	SidechainEndpoint    url.URL
	ContractRegistryAddr common.Address
	BlocksBatchSize      uint64
	MasterchainGasPrice  *big.Int
}

func NewDefaultConfig() (*Config, error) {
	endpoint, err := url.Parse(defaultMasterchainEndpoint)
	if err != nil {
		return nil, err
	}

	sidechainEndpoint, err := url.Parse(defaultSidechainEndpoint)
	if err != nil {
		return nil, err
	}

	return &Config{
		Endpoint:             *endpoint,
		SidechainEndpoint:    *sidechainEndpoint,
		ContractRegistryAddr: common.HexToAddress(defaultContractRegistryAddr),
	}, nil
}

func (m *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg struct {
		MasterchainEndpoint  string         `yaml:"endpoint"`
		SidechainEndpoint    string         `yaml:"sidechain_endpoint"`
		ContractRegistryAddr common.Address `yaml:"contract_registry"`
		BlocksBatchSize      uint64         `yaml:"blocks_batch_size" default:"500"`
		MasterchainGasPrice  GasPrice       `yaml:"masterchain_gas_price"`
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
	m.BlocksBatchSize = cfg.BlocksBatchSize
	m.MasterchainGasPrice = cfg.MasterchainGasPrice.Int

	return nil
}
