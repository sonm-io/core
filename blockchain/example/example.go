package main

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"

	//"github.com/sonm-io/core/proto"
)

const testPass = "QWEpoi123098"

func main() {
	var err error

	ks := accounts.NewIdentity("sonm-test-keystore")

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

	bch, err := blockchain.NewAPI(nil, nil)
	if err != nil {
		log.Fatal(bch)
		return
	}

	var tx *types.Transaction

	//tx, err = bch.Approve(prv, tsc.DealsAddress, big.NewInt(10000))
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}

	deal := sonm.Deal{
		BuyerID:           "0x41ba7e0e1e661f7114f2f05afd2536210c2ed351",
		SupplierID:        "0x41ba7e0e1e661f7114f2f05afd2536210c2ed352",
		SpecificationHash: "1234567890",
		Price:             "10000",
		WorkTime:          60,
	}

	tx, err = bch.OpenDeal(prv, &deal)
	if err != nil {
		log.Fatalln(err)
		return
	}

	//tx, err := bch.AcceptDeal(prv, big.NewInt(2))
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}

	//tx, err = bch.CloseDeal(prv, big.NewInt(1))
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}

	log.Println(tx)

}
