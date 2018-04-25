package main

import (
	"context"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

func main() {
	prv, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalln(err)
		return
	}

	api, err := blockchain.NewAPI()
	if err != nil {
		log.Fatalln(err)
		return
	}

	order := &sonm.Order{
		OrderType:      sonm.OrderType_ASK,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       crypto.PubkeyToAddress(prv.PublicKey).Hex(),
		CounterpartyID: "0x0",
		Duration:       3600 - 50 + (rand.Uint64() % 100),
		Price:          sonm.NewBigIntFromInt(1000 + rand.Int63n(1000)),
		Netflags:       sonm.NetflagsToUint([3]bool{true, true, (rand.Int() % 2) == 0}),
		IdentityLevel:  sonm.IdentityLevel_ANONYMOUS,
		Blacklist:      "0x0",
		Tag:            []byte("00000"),
		Benchmarks: &sonm.Benchmarks{
			Values: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}

	res := <- api.PlaceOrder(
		context.Background(),
		prv,
		order,
	)
	if res.Err != nil {
		log.Fatalln(res.Err.Error())
		return
	}

	log.Println(res.Order.Id)
	ordId, err := util.ParseBigInt(res.Order.Id)
	if err != nil {
		log.Fatalln("Cannot cast")
		return
	}
	err = <- api.CancelOrder(context.Background(), prv, ordId)

	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("canceled")
}
