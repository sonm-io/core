package commands

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/cmd/cli/config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonOutputForOrder(t *testing.T) {
	buf := initRootCmd(t, "", config.OutputModeJSON)

	bigVal, _ := pb.NewBigIntFromString("1000000000000000000000000000")
	printOrdersList(rootCmd, []*pb.Order{{
		Price: bigVal,
	},
	})

	out := buf.String()
	assert.Equal(t, "{\"orders\":[{\"price\":\"1000000000000000000000000000\"}]}\r\n", out,
		"price must be serialized as string, not `abs` and `neg` parts of pb.BigInt")
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

	deal := &pb.Deal{
		Status:      pb.DealStatus_DEAL_CLOSED,
		Id:          pb.NewBigIntFromInt(1488),
		ConsumerID:  pb.NewEthAddress(common.HexToAddress("0x111")),
		SupplierID:  pb.NewEthAddress(common.HexToAddress("0x222")),
		Price:       pb.NewBigIntFromInt(1e18),
		StartTime:   &pb.Timestamp{Seconds: 0},
		EndTime:     &pb.Timestamp{Seconds: 0},
		LastBillTS:  &pb.Timestamp{Seconds: 0},
		AskID:       pb.NewBigIntFromInt(1),
		BidID:       pb.NewBigIntFromInt(2),
		TotalPayout: pb.NewBigIntFromInt(5),
	}

	buf := initRootCmd(t, "", config.OutputModeSimple)
	cfg = &config.Config{Eth: accounts.EthConfig{Passphrase: "test"}}

	printDealInfo(rootCmd, &pb.DealInfoReply{Deal: deal}, nil, suppressWarnings)

	assert.Contains(t, buf.String(), "Duration:     0s")
}

func TestFlags(t *testing.T) {
	assert.False(t, printerFlags(printEverything).WarningSuppressed())
	assert.True(t, printerFlags(suppressWarnings).WarningSuppressed())
}
