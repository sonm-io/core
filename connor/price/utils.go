package price

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/params"
)

const (
	retryCount   = 3
	retryTimeout = 1 * time.Second

	whatToMineURL = "https://whattomine.com/coins.json"
	zcashWtmID    = 166
	ethWtmID      = 151
	moneroEtmID   = 101
)

type wtmCoin struct {
	ID          int     `json:"id"`
	Difficulty  float64 `json:"difficulty"`
	BlockReward float64 `json:"block_reward"`
	Nethash     int     `json:"nethash"`
	// warn: somewhere it is string, somewhere int
	// BlockTime   string  `json:"block_time"`
}

type wtmResponse struct {
	Coins map[string]*wtmCoin `json:"coins"`
}

func getTokenParamsFromWTM(id int) (*wtmCoin, error) {
	data, err := FetchURLWithRetry(whatToMineURL)
	if err != nil {
		return nil, err
	}

	resp := &wtmResponse{}
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal token params response: %v", err)
	}

	for _, coin := range resp.Coins {
		if coin.ID == id {
			return coin, nil
		}
	}

	return nil, fmt.Errorf("cannot find coin with id %d in response", id)
}

type coinMarketCapResponse struct {
	PriceUSD string `json:"price_usd"`
}

func getPriceFromCMC(url string) (*big.Int, error) {
	data, err := FetchURLWithRetry(url)
	if err != nil {
		return nil, err
	}

	var resp []*coinMarketCapResponse = nil
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal price response: %v", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("coinMarketCap returns an empty array")
	}

	v, err := strconv.ParseFloat(resp[0].PriceUSD, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot prase token price: %v", err)
	}

	price, _ := big.NewFloat(0).Mul(big.NewFloat(v), big.NewFloat(params.Ether)).Int(nil)
	return price, nil
}

func FetchURLWithRetry(url string) ([]byte, error) {
	body, err := fetchOnce(url)
	if err != nil {
		for i := 0; i < retryCount; i++ {
			time.Sleep(retryTimeout)

			body, err = fetchOnce(url)
			if err != nil {
				continue
			}

			return body, nil
		}
		return nil, fmt.Errorf("http connection retries exceeded, cannot perform http request to %s", url)
	}
	return body, nil
}

func fetchOnce(url string) ([]byte, error) {
	// todo: ctxhttp.Get()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request for %s returns status %d", url, resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
