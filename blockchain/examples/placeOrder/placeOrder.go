package main

import (
	"context"
	"log"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

const (
	hexKey = "a5dd45e0810ca83e21f1063e6bf055bd13544398f280701cbfda1346bcf3ae64"
)

func main() {
	prv, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Fatalln(err)
		return
	}

	api, err := blockchain.NewAPI()
	if err != nil {
		log.Fatalln(err)
		return
	}

	balance, err := api.SideToken().BalanceOf(context.Background(), crypto.PubkeyToAddress(prv.PublicKey))
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Balance: ", balance)

	allowance, err := api.SideToken().AllowanceOf(context.Background(), crypto.PubkeyToAddress(prv.PublicKey), blockchain.MarketAddr())
	if err != nil {
		log.Fatalln(err)
		return
	}

	log.Println("Allowance: ", allowance)

	price := sonm.NewBigIntFromInt(1)

	if balance.Cmp(price.Unwrap()) < 0 {
		log.Fatalln("lack of balance")
		return
	}

	order := &sonm.Order{
		OrderType:      sonm.OrderType_BID,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       sonm.NewEthAddress(crypto.PubkeyToAddress(prv.PublicKey)),
		CounterpartyID: sonm.NewEthAddress(common.HexToAddress("0x0")),
		Duration:       3600 - 50 + (rand.Uint64() % 100),
		Price:          price,
		Netflags:       sonm.NetflagsToUint([3]bool{true, true, (rand.Int() % 2) == 0}),
		IdentityLevel:  sonm.IdentityLevel_ANONYMOUS,
		Blacklist:      "0x0",
		Tag:            []byte("00000"),
		Benchmarks: &sonm.Benchmarks{
			Values: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}

	res := <-api.Market().PlaceOrder(
		context.Background(),
		prv,
		order,
	)
	if res.Err != nil {
		log.Fatalln(res.Err.Error())
		return
	}

	log.Println(res.Order.Id.Unwrap().String())
	ordId, err := util.ParseBigInt(res.Order.Id.Unwrap().String())
	if err != nil {
		log.Fatalln("Cannot cast")
		return
	}
	log.Println("Canceling")

	err = <-api.Market().CancelOrder(context.Background(), prv, ordId)

	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("canceled")
}
