package tests

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

type Account struct {
	Key     []byte
	Address common.Address
	TxOpts  *bind.TransactOpts
}

type TestRpc struct {
	RpcCli   *rpc.Client
	EthCli   *ethclient.Client
	Accounts []Account
}

// globals
var trpc TestRpc

// constants
var (
	numberOfTestAccounts      = 64 // up to 255 accounts is supported
	numberOfConnectionRetries = 7
	connectionTimeout         = 1 * time.Second
)

func startTestRPC(binPath string, nOfAccounts int) (error, func()) {
	var args []string
	for i := 0; i < numberOfTestAccounts; i++ {
		key := bytes.Repeat([]byte{(byte)(i + 1)}, 32)
		arg := "--account=0x" + hex.EncodeToString(key) +
			",1234" + strings.Repeat("0", 18)
		args = append(args, arg)

		txOpts := bind.NewKeyedTransactor(crypto.ToECDSAUnsafe(key))
		trpc.Accounts = append(trpc.Accounts, Account{key, txOpts.From, txOpts})
	}

	testrpc := exec.Command(os.Getenv("TESTRPC"), args...)

	// set process group id to kill forked process with all its subprocesses
	testrpc.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return testrpc.Start(), func() {
		syscall.Kill(-testrpc.Process.Pid, syscall.SIGTERM)
		testrpc.Wait()
	}
}

func runTestsWithTestrpc(testrpcPath string, m *testing.M) error {
	err, wait := startTestRPC(testrpcPath, numberOfTestAccounts)
	if err != nil {
		return errors.Wrap(err, "failed to start testrpc")
	}
	defer wait()

	trpc.RpcCli, _ = rpc.Dial("http://localhost:8545")
	trpc.EthCli = ethclient.NewClient(trpc.RpcCli)

	for i := 0; i < numberOfConnectionRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
		defer cancel()
		err = trpc.RpcCli.CallContext(ctx, &trpc.Accounts, "eth_accounts")
		<-ctx.Done()
		if len(trpc.Accounts) > 0 {
			break
		}
	}

	if res := m.Run(); res != 0 {
		return errors.New("failed")
	}
	return nil
}
