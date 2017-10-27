package utils

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
)

func InitEthClient(ethEndpoint string) (*ethclient.Client, error) {
	ethClient, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return ethClient, err
}
