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

	api, err := blockchain.NewAPI(blockchain.WithBlockConfirmations(0))
	if err != nil {
		log.Fatalln(err)
		return
	}

	balance, err := api.MasterchainToken().BalanceOf(context.Background(), crypto.PubkeyToAddress(key.PublicKey))
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Balance: ", balance)

	allowance, err := api.MasterchainToken().AllowanceOf(context.Background(), crypto.PubkeyToAddress(key.PublicKey), blockchain.GatekeeperMasterchainAddr())
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
		err = api.MasterchainToken().Approve(context.TODO(), key, blockchain.GatekeeperMasterchainAddr(), balance)
		if err != nil {
			log.Println("approve: ", err)
		}
	}

	err = api.MasterchainGate().PayIn(context.TODO(), key, price)
	if err != nil {
		log.Fatalln("payin: ", err)
		return
	}

	log.Println("done.")
}
