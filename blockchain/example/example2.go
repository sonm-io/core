package main

import (
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/tsc"
	"log"
	"math/big"
)

const testPass = ""

func main() {
	var err error

	ks := accounts.NewIdentity(accounts.GetDefaultKeystoreDir())

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

	result, err := blockchain.GetHubDeals("0x41ba7e0e1e661f7114f2f05afd2536210c2ed351")
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("HubDeals: ", result)

	result, err = blockchain.GetClientDeals("0x41ba7e0e1e661f7114f2f05afd2536210c2ed351")
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("ClientDeals: ", result)

	dealAmount, err := blockchain.GetDealAmount()
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("dealAmount: ", dealAmount)

	dealInfo, err := blockchain.GetDealInfo(big.NewInt(4))
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("dealInfo: ", dealInfo)
	log.Println("dealInfo-SpecHash: ", dealInfo.SpecificationHash)
	log.Println("dealInfo-Client: ", dealInfo.Client)
	log.Println("dealInfo-Hub: ", dealInfo.Hub)
	log.Println("dealInfo-Price: ", dealInfo.Price.String())
	log.Println("dealInfo-Status: ", dealInfo.Status.String())
	log.Println("dealInfo-StartTime: ", dealInfo.StartTime.String())
	log.Println("dealInfo-WorkTime: ", dealInfo.WorkTime.String())
	log.Println("dealInfo-EndTime: ", dealInfo.EndTime.String())

	balance, err := blockchain.BalanceOf("0x41ba7e0e1e661f7114f2f05afd2536210c2ed351")
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("Balance: ", balance)

	allowance, err := blockchain.AllowanceOf("0x41ba7e0e1e661f7114f2f05afd2536210c2ed351", tsc.DealsAddress)
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("Allowance: ", allowance)

}
