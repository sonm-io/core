package commands

import (
	"testing"

	"sort"

	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestXXX(t *testing.T) {
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
