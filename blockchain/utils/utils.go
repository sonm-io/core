package utils

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
)

func InitEthClient() (*ethclient.Client, error){
	ethClient, err := ethclient.Dial("https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	return ethClient, err
}