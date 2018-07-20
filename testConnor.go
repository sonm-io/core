package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

func main() {
	pricePerSec := 0.00000008843452498661306
	realV := big.NewFloat(pricePerSec)
	price := realV.Mul(realV, big.NewFloat(params.Ether))
	actualPrice, _ := price.Int(nil)

	if actualPrice.Cmp(big.NewInt(0)) == 0 {
		log.Printf("Pidor")
	}
	fmt.Printf("price %s",actualPrice)
}
