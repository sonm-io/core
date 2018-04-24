package main

import (
	"context"
	"log"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
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

	chPlace := api.PlaceOrder(
		context.Background(),
		prv,
		order,
	)
	res := <-chPlace
	if res.Err != nil {
		log.Println(res.Err.Error())
		return
	}

	log.Println(res.Order.Id)
	ordId, cast := big.NewInt(0).SetString(res.Order.Id, 10)
	if !cast {
		log.Println("Cannot cast")
		return
	}
	chCancel := api.CancelOrder(context.Background(), prv, ordId)

	resCancel := <-chCancel
	if resCancel != nil {
		log.Println(resCancel)
	}
	log.Println("canceled")
}
