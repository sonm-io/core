package structs

import (
	"errors"
	pb "github.com/sonm-io/core/proto"
)

func ParseNetworkType(ty string) (pb.NetworkType, error) {
	typeID, ok := pb.NetworkType_value[ty]
	if !ok {
		return pb.NetworkType_NO_NETWORK, errors.New("unknown network type")
	}

	return pb.NetworkType(typeID), nil
}
