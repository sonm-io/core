package tc

import (
	"fmt"
)

type Protocol int

func (m Protocol) String() string {
	switch m {
	case ProtoAll:
		return "all"
	case ProtoIP:
		return "ip"
	default:
		return fmt.Sprintf("unknown protocol: %d", m)
	}
}
