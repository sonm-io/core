package datastruct

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Node holds information about a Node.
type Node struct {
	EthAddr common.Address `json:"eth_addr"`
	IpAddr  []string       `json:"ip_addr"`
	TS      time.Time      `json:"time"`
}
