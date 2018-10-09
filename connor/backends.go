package connor

import (
	"github.com/sonm-io/core/connor/antifraud"
	"github.com/sonm-io/core/connor/price"
	"github.com/sonm-io/core/connor/types"
)

type backends struct {
	corderFactory    types.CorderFactory
	dealFactory      types.DealFactory
	priceProvider    price.Provider
	processorFactory antifraud.ProcessorFactory
}

func NewBackends(cfg *Config) *backends {
	return &backends{
		processorFactory: antifraud.NewProcessorFactory(&cfg.AntiFraud),
		corderFactory:    types.NewCorderFactory(cfg.Container.Tag, cfg.Market.benchmarkID, cfg.Market.Counterparty),
		dealFactory:      types.NewDealFactory(cfg.Market.benchmarkID),
		priceProvider:    cfg.PriceSource.Init(cfg.Market.PriceControl.Marginality),
	}
}
