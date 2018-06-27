package nppc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type ResourceID struct {
	Protocol string
	Addr     common.Address
}

func (m ResourceID) String() string {
	return fmt.Sprintf("%s://%s", m.Protocol, m.Addr.Hex())
}
