package datastruct

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Node holds information about a Node.
type Node struct {
	EthAddr common.Address
	IpAddr  []string
	TS      time.Time
}
