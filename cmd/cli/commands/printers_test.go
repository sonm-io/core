package commands

import (
	"testing"

	"github.com/sonm-io/core/cmd/cli/config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestJsonOutputForOrder(t *testing.T) {
	buf := initRootCmd(t, "", config.OutputModeJSON)

	bigVal, _ := pb.NewBigIntFromString("1000000000000000000000000000")
	printSearchResults(rootCmd, []*pb.Order{{
		Price: bigVal,
	},
	})

	out := buf.String()
	assert.Equal(t, "{\"orders\":[{\"price\":\"1000000000000000000000000000\"}]}\r\n", out,
		"price must be serialized as string, not `abs` and `neg` parts of pb.BigInt")
}

func TestDealInfoWithZeroDuration(t *testing.T) {
	deal := &pb.Deal{
		Status:     pb.DealStatus_DEAL_CLOSED,
		Id:         "1488",
		ConsumerID: "0x111",
		SupplierID: "0x222",
		Price:      pb.NewBigIntFromInt(1e18),
		StartTime:  &pb.Timestamp{Seconds: 0},
		EndTime:    &pb.Timestamp{Seconds: 0},
	}

	buf := initRootCmd(t, "", config.OutputModeSimple)
	printDealInfo(rootCmd, deal)

	assert.Contains(t, buf.String(), "Duraton:  0s")
}
