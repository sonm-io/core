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

func ParseGPUCount(ty string) (pb.GPUCount, error) {
	typeID, ok := pb.GPUCount_value[ty]
	if !ok {
		return pb.GPUCount_NO_GPU, errors.New("unknown gpu count")
	}

	count := pb.GPUCount(typeID)
	if count == pb.GPUCount_SINGLE_GPU {
		return pb.GPUCount_NO_GPU, ErrUnsupportedSingleGPU
	}

	return count, nil
}
