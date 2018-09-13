package structs

import (
	"github.com/sonm-io/core/proto"
)

type NetworkSpec struct {
	*sonm.NetworkSpec
	NetID string
}
