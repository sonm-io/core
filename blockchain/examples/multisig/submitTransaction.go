package main

import (
	"context"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/source/api"
)

func main() {
	// Blockchain web3 json rpc endpoint
	// https://sidechain.livenet.sonm.com
	// https://sidechain-dev.sonm.com - for testnet
	blockchainEndpoint := ""

	// msAddress where transaction would be placed
	msAddress := common.HexToAddress("")

	// multisig-transaction destination
	dest := common.HexToAddress("")

	// multisig-transaction value
	value := big.NewInt(0)

	// multisig-transaction data
	data := common.FromHex("")

	// private key in hex format
	key, err := crypto.HexToECDSA("")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	opts := bind.NewKeyedTransactor(key)
	opts.Context = context.TODO()
	opts.GasLimit = 2000000
	opts.GasPrice = big.NewInt(0)

	client, err := blockchain.NewClient(blockchainEndpoint)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	ms, err := api.NewMultiSigWallet(msAddress, client)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	tx, err := ms.SubmitTransaction(opts, dest, value, data)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	rec, err := blockchain.WaitTransactionReceipt(context.TODO(), client, 1, 1, tx)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	log.Println(rec)
}
