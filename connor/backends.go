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
