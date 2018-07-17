package watchers

import (
	"context"
	"math/big"
)

const (
	coinMarketCapTicker     = "https://api.coinmarketcap.com/v1/ticker/"
	coinMarketCapSonmTicker = coinMarketCapTicker + "sonm/"
	cryptoCompareCoinData   = "https://www.cryptocompare.com/api/data/coinsnapshotfullbyid/?id="
)

// Watcher is watching for external resources updates
type Watcher interface {
	Update(ctx context.Context) error
}

type PriceWatcher interface {
	Watcher
	GetPrice() float64
}

type TokenWatcher interface {
	Watcher
	GetTokenData(symbol string) (*TokenParameters, error)
}

type PoolWatcher interface {
	Watcher
	GetData(addr string) (*ReportedHashrate, error)
}

type BudgetWatcher interface {
	Watcher
	GetBalance() *big.Int
}
