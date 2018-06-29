package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
)

func main() {
	const hexKey = "a5dd45e0810ca83e21f1063e6bf055bd13544398f280701cbfda1346bcf3ae64"
	price := big.NewInt(100)

	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Fatalln(err)
		return
	}

	api, err := blockchain.NewAPI()
	if err != nil {
		log.Fatalln(err)
		return
	}

	balance, err := api.SidechainToken().BalanceOf(context.Background(), crypto.PubkeyToAddress(key.PublicKey))
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Balance: ", balance)

	allowance, err := api.SidechainToken().AllowanceOf(context.Background(), crypto.PubkeyToAddress(key.PublicKey), blockchain.GatekeeperSidechainAddr())
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Allowance: ", allowance)

	if balance.Cmp(price) < 0 {
		log.Fatalln("lack of balance")
		return
	}

	if allowance.Cmp(price) < 0 {
		log.Println("lack of allowance, set new")
		err = api.SidechainToken().Approve(context.TODO(), key, blockchain.GatekeeperSidechainAddr(), balance)
		if err != nil {
			log.Fatalln("approve: ", err)
			return
		}
	}

	err = api.SidechainGate().PayIn(context.TODO(), key, price)
	if err != nil {
		log.Fatalln("payin: ", err)
		return
	}

	log.Println("done.")
}
