package contracts

import (
	"context"
	"fmt"
	"github.com/sonm-io/core/contracts/api"
	"github.com/stretchr/testify/assert"
	"math/big"
	"os"
	"testing"
)

func TestDeployHubWallet(t *testing.T) {
	tokenOwner := trpc.Accounts[0]
	hubOwner := trpc.Accounts[1]
	unused := trpc.Accounts[42]

	tokenAddr, _, _, err :=
		api.DeploySonmDummyToken(tokenOwner.TxOpts, trpc.EthCli, tokenOwner.Address)
	assert.Empty(t, err, "should deploy SonmDummyToken successfully")

	_, _, hWallet, err :=
		api.DeployHubWallet(
			hubOwner.TxOpts,
			trpc.EthCli,
			hubOwner.Address,
			unused.Address, // DAO
			unused.Address, // whiteList
			tokenAddr)
	assert.Empty(t, err, "should deploy HubWallet successfully")

	phase, err := hWallet.CurrentPhase(nil)
	assert.Empty(t, err, "should call CurrentPhase sucessfully")
	assert.EqualValues(t, 2, phase, "should have currentPase = Idle after creation")
}

func TestAccounts(t *testing.T) {
	balance, _ := trpc.EthCli.BalanceAt(context.Background(), trpc.Accounts[0].Address, nil)
	assert.True(t, balance.Cmp(&big.Int{}) > 0, "account balance")
}

func TestMain(m *testing.M) {
	err := runTestsWithTestrpc(os.Getenv("TESTRPC"), m)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	} else {
		os.Exit(0)
	}
}
