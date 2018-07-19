package blockchain

import (
	"fmt"
	"math/big"
	"strings"
)

type Unit struct {
	*big.Int
}

var (
	Wei    = Unit{Int: big.NewInt(1)}
	KWei   = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil)}
	MWei   = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)}
	GWei   = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)}
	Szabo  = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)}
	Finney = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil)}
	Ether  = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)}
	KEther = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(21), nil)}
	MEther = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(24), nil)}
	GEther = Unit{Int: new(big.Int).Exp(big.NewInt(10), big.NewInt(27), nil)}
)

func UnitFromString(v string) (Unit, error) {
	switch strings.ToLower(v) {
	case "wei":
		return Wei, nil
	case "kwei":
		return KWei, nil
	case "mwei":
		return MWei, nil
	case "gwei":
		return GWei, nil
	case "szabo":
		return Szabo, nil
	case "finney":
		return Finney, nil
	case "ether":
		return Ether, nil
	case "kether":
		return KEther, nil
	case "mether":
		return MEther, nil
	case "gether":
		return GEther, nil
	default:
		return Unit{}, fmt.Errorf("unknown ETH unit: %s", v)
	}
}
