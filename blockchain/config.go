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
	Version              uint
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
		Version:              defaultVersion,
	}, nil
}

func (m *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg struct {
		MasterchainEndpoint  string         `yaml:"endpoint"`
		SidechainEndpoint    string         `yaml:"sidechain_endpoint"`
		ContractRegistryAddr common.Address `yaml:"contract_registry"`
		BlocksBatchSize      uint64         `yaml:"blocks_batch_size"`
		MasterchainGasPrice  GasPrice       `yaml:"masterchain_gas_price"`
		Version              uint           `yaml:"version"`
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

	if cfg.Version == 0 {
		cfg.Version = defaultVersion
	}

	endpoint, err := url.Parse(cfg.MasterchainEndpoint)
	if err != nil {
		return err
	}

	sidechainEndpoint, err := url.Parse(cfg.SidechainEndpoint)
	if err != nil {
		return err
	}

	if cfg.BlocksBatchSize == 0 {
		cfg.BlocksBatchSize = 500
	}

	m.Endpoint = *endpoint
	m.SidechainEndpoint = *sidechainEndpoint
	m.ContractRegistryAddr = cfg.ContractRegistryAddr
	m.BlocksBatchSize = cfg.BlocksBatchSize
	m.MasterchainGasPrice = cfg.MasterchainGasPrice.Int
	m.Version = cfg.Version

	return nil
}
