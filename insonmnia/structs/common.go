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

func ParseOrderType(ty string) (pb.OrderType, error) {
	typeID, ok := pb.OrderType_value[ty]
	if !ok {
		return pb.OrderType_ANY, errors.New("unknown order type")
	}

	return pb.OrderType(typeID), nil
}
