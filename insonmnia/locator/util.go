package locator

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/net/context"

	"github.com/sonm-io/core/util"
)

func ExtractEthAddr(ctx context.Context) (common.Address, error) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return common.Address{}, status.Error(codes.DataLoss, "failed to get peer from ctx")
	}

	switch info := pr.AuthInfo.(type) {
	case util.EthAuthInfo:
		return info.Wallet, nil
	default:
		return common.Address{}, status.Error(codes.Unauthenticated, "wrong AuthInfo type")
	}
}
