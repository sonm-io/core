package main

import (
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"log"
	"math/big"
)

const testPass = ""

func main() {
	var err error

	ks := accounts.NewIdentity("/tmp/sonm-test-keystore")

	err = ks.Open(testPass)
	if err != nil {
		if err == accounts.ErrWalletIsEmpty {
			err = ks.New(testPass)
			if err != nil {
				log.Fatal(err)
			}
		}
		log.Fatal(err)
	}

	prv, err := ks.GetPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	bch, err := blockchain.NewBlockchainAPI(prv, nil)
	if err != nil {
		log.Fatal(bch)
		return
	}

	tx, err := bch.OpenDeal("0x41ba7e0e1e661f7114f2f05afd2536210c2ed351", "0x41ba7e0e1e661f7114f2f05afd2536210c2ed351", big.NewInt(1236782361542612), big.NewInt(10000), big.NewInt(3600))
	if err != nil {
		log.Fatalln(err)
		return
	}

	//tx, err := bch.AcceptDeal(*big.NewInt(3))
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}

	//tx, err := bch.CloseDeal(big.NewInt(4))
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}

	log.Println(tx)

}
