package commands

import (
	"sort"
	"testing"

	"github.com/sonm-io/core/cmd/cli/config"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestOrdersSort(t *testing.T) {
	in := &pb.GetProcessingReply{
		Orders: map[string]*pb.GetProcessingReply_ProcessedOrder{
			"ccc": {
				Id:        "ccc",
				Status:    1,
				Extra:     "",
				Timestamp: &pb.Timestamp{Seconds: 111},
			},
			"aaa": {
				Id:        "aaa",
				Status:    1,
				Extra:     "",
				Timestamp: &pb.Timestamp{Seconds: 333},
			},
			"bbb": {
				Id:        "bbb",
				Status:    1,
				Extra:     "",
				Timestamp: &pb.Timestamp{Seconds: 222},
			},
		},
	}

	ls := make([]*pb.GetProcessingReply_ProcessedOrder, 0, len(in.GetOrders()))
	for _, item := range in.GetOrders() {
		ls = append(ls, item)
	}

	sort.Sort(handlerByTime(ls))

	assert.Equal(t, "ccc", ls[0].Id)
	assert.Equal(t, "bbb", ls[1].Id)
	assert.Equal(t, "aaa", ls[2].Id)
}

func TestJsonOutputForOrder(t *testing.T) {
	buf := initRootCmd(t, "", config.OutputModeJSON)

	bigVal, _ := pb.NewBigIntFromString("1000000000000000000000000000")
	printSearchResults(rootCmd, []*pb.Order{{
		PricePerSecond: bigVal,
	},
	})

	out := buf.String()
	assert.Equal(t, "{\"orders\":[{\"pricePerSecond\":\"1000000000000000000000000000\"}]}\r\n", out,
		"price must be serialized as string, not `abs` and `neg` parts of pb.BigInt")
}
