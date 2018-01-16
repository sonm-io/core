package main

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
)

const testPass = "any"

func main() {
	var err error

	ks := accounts.NewIdentity("keys")

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

	deal := pb.Deal{
		BuyerID:           "0x8125721c2413d99a33e351e1f6bb4e56b6b633fd",
		SupplierID:        "0x8125721c2413d99a33e351e1f6bb4e56b6b633fd",
		SpecificationHash: "1234567890",
		Price:             pb.NewBigIntFromInt(100),
		WorkTime:          60,
	}

	dealId, err := bch.OpenDealPending(context.Background(), prv, &deal, time.Duration(180*time.Second))

	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("DealId: ", dealId)

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
