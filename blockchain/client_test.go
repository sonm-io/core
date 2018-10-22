package blockchain

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBlockNumberWithout0x(t *testing.T) {
	raw := []byte(`{
"blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037",
"blockNumber":"ff",
"contractAddress":null,
"cumulativeGasUsed":"0x81ce3",
"from":"0x4ac11b6ed0f118414db41b41dade342368f925ca",
"gasUsed":"0x81ce3",
"logs":[
{"address":"0x6c88e07debd749476636ac4841063130df6c62bf",
"topics":["0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d","0x0000000000000000000000000000000000000000000000000000000000067b69"],
"data":"0x",
"blockNumber":"0x359f08",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0","blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037","logIndex":"0x0","removed":false}],
"logsBloom":"0x00000000000000000000000000000008000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000020000000001000000000000000000400000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000",
"status":"0x1",
"to":"0x6c88e07debd749476636ac4841063130df6c62bf",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0"
}`)

	rec := &Receipt{
		Receipt:     &types.Receipt{},
		BlockNumber: 0,
	}
	err := rec.UnmarshalJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, int64(255), rec.BlockNumber)
}

func TestParseBlockNumberWith0x(t *testing.T) {
	raw := []byte(`{
"blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037",
"blockNumber":"0x359f08",
"contractAddress":null,
"cumulativeGasUsed":"0x81ce3",
"from":"0x4ac11b6ed0f118414db41b41dade342368f925ca",
"gasUsed":"0x81ce3",
"logs":[
{"address":"0x6c88e07debd749476636ac4841063130df6c62bf",
"topics":["0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d","0x0000000000000000000000000000000000000000000000000000000000067b69"],
"data":"0x",
"blockNumber":"0x359f08",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0","blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037","logIndex":"0x0","removed":false}],
"logsBloom":"0x00000000000000000000000000000008000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000020000000001000000000000000000400000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000",
"status":"0x1",
"to":"0x6c88e07debd749476636ac4841063130df6c62bf",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0"
}
`)

	rec := &Receipt{
		Receipt:     &types.Receipt{},
		BlockNumber: 0,
	}
	err := rec.UnmarshalJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, int64(3514120), rec.BlockNumber)
}

func TestParseParseTo(t *testing.T) {
	raw := []byte(`{
"blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037",
"blockNumber":"0x359f08",
"contractAddress":null,
"cumulativeGasUsed":"0x81ce3",
"from":"0x4ac11b6ed0f118414db41b41dade342368f925ca",
"gasUsed":"0x81ce3",
"logs":[
{"address":"0x6c88e07debd749476636ac4841063130df6c62bf",
"topics":["0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d","0x0000000000000000000000000000000000000000000000000000000000067b69"],
"data":"0x",
"blockNumber":"0x359f08",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0","blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037","logIndex":"0x0","removed":false}],
"logsBloom":"0x00000000000000000000000000000008000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000020000000001000000000000000000400000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000",
"status":"0x1",
"to":"0x6c88e07debd749476636ac4841063130df6c62bf",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0"
}`)

	rec := &Receipt{
		Receipt:     &types.Receipt{},
		BlockNumber: 0,
	}
	err := rec.UnmarshalJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, strings.ToLower(rec.To.String()), "0x6c88e07debd749476636ac4841063130df6c62bf")
}

func TestParseParseToWithNull(t *testing.T) {
	raw := []byte(`{
"blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037",
"blockNumber":"0x359f08",
"contractAddress":null,
"cumulativeGasUsed":"0x81ce3",
"from":"0x4ac11b6ed0f118414db41b41dade342368f925ca",
"gasUsed":"0x81ce3",
"logs":[
{"address":"0x6c88e07debd749476636ac4841063130df6c62bf",
"topics":["0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d","0x0000000000000000000000000000000000000000000000000000000000067b69"],
"data":"0x",
"blockNumber":"0x359f08",
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0","blockHash":"0x1929947310bd7aae9509b99f8986297969e7450e116712c85d9c77a959bb8037","logIndex":"0x0","removed":false}],
"logsBloom":"0x00000000000000000000000000000008000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000020000000001000000000000000000400000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000",
"status":"0x1",
"to": null,
"transactionHash":"0xf79deb72c6eea1d89490cd4d4706bfb50e6d96700021ed79cee0238012b072d2",
"transactionIndex":"0x0"
}`)

	rec := &Receipt{
		Receipt:     &types.Receipt{},
		BlockNumber: 0,
	}
	err := rec.UnmarshalJSON(raw)
	require.NoError(t, err)
	assert.Equal(t, rec.To.String(), "0x0000000000000000000000000000000000000000")
}
