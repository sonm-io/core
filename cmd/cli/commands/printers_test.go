package commands

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonOutputForOrder(t *testing.T) {
	cfg = &config.Config{OutFormat: config.OutputModeJSON}
	buf := initRootCmd(t, "")

	bigVal, _ := sonm.NewBigIntFromString("1000000000000000000000000000")
	printOrdersList(rootCmd, []*sonm.Order{{
		Price: bigVal,
	},
	})

	out := buf.String()
	assert.Equal(t, "{\"orders\":[{\"price\":\"1000000000000000000000000000\"}]}\r\n", out,
		"price must be serialized as string, not `abs` and `neg` parts of sonm.BigInt")
}

func TestDealInfoWithZeroDuration(t *testing.T) {
	keydir := os.TempDir()
	defer os.Remove(keydir)

	var err error
	keystore, err = accounts.NewMultiKeystore(&accounts.KeystoreConfig{
		KeyDir:      keydir,
		PassPhrases: make(map[string]string),
	}, accounts.NewStaticPassPhraser("test"))
	require.NoError(t, err)

	generatedKey, err := keystore.GenerateWithPassword("test")
	require.NoError(t, err)

	err = keystore.SetDefault(crypto.PubkeyToAddress(generatedKey.PublicKey))
	require.NoError(t, err)

	deal := &sonm.Deal{
		Status:      sonm.DealStatus_DEAL_CLOSED,
		Id:          sonm.NewBigIntFromInt(1488),
		ConsumerID:  sonm.NewEthAddress(common.HexToAddress("0x111")),
		SupplierID:  sonm.NewEthAddress(common.HexToAddress("0x222")),
		Price:       sonm.NewBigIntFromInt(1e18),
		StartTime:   &sonm.Timestamp{Seconds: 0},
		EndTime:     &sonm.Timestamp{Seconds: 0},
		LastBillTS:  &sonm.Timestamp{Seconds: 0},
		AskID:       sonm.NewBigIntFromInt(1),
		BidID:       sonm.NewBigIntFromInt(2),
		TotalPayout: sonm.NewBigIntFromInt(5),
	}

	cfg = &config.Config{OutFormat: config.OutputModeSimple, Eth: accounts.EthConfig{Passphrase: "test"}}
	buf := initRootCmd(t, "")

	info := &ExtendedDealInfo{
		DealInfoReply: &sonm.DealInfoReply{Deal: deal},
	}
	printDealInfo(rootCmd, info, suppressWarnings)

	assert.Contains(t, buf.String(), "Duration:     0s")
}

func TestFlags(t *testing.T) {
	assert.False(t, printerFlags(printEverything).WarningSuppressed())
	assert.True(t, printerFlags(suppressWarnings).WarningSuppressed())
}

func TestExpensesPerHour(t *testing.T) {
	my := common.HexToAddress("0x928cA7817FE2eBAC8C41e9dEF8EA6c09ffbd385A")
	other := common.HexToAddress("0xEAed1BFb645Dd9ca85a062B4e5eF34857aecdd4E")
	deals := []*pb.Deal{
		{Price: pb.NewBigIntFromInt(100), SupplierID: pb.NewEthAddress(my)},
		{Price: pb.NewBigIntFromInt(150), SupplierID: pb.NewEthAddress(my)},
		{Price: pb.NewBigIntFromInt(200), SupplierID: pb.NewEthAddress(my)},

		{Price: pb.NewBigIntFromInt(123), SupplierID: pb.NewEthAddress(other)},
		{Price: pb.NewBigIntFromInt(450), SupplierID: pb.NewEthAddress(other)},
	}

	asks, bids := dealsExpensesPerHour(my, deals)
	assert.Equal(t, pb.NewBigIntFromInt(450*3600), asks)
	assert.Equal(t, pb.NewBigIntFromInt(573*3600), bids)
}
