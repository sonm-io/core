package sonm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBidOrderValidate(t *testing.T) {
	bid := &BidOrder{Tag: "this-string-is-too-long-for-tag-value"}
	err := bid.Validate()
	require.Error(t, err)

	bid.Tag = "short-and-valid"
	err = bid.Validate()
	require.NoError(t, err)
}
