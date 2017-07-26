package util

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"net"
)

func ToPubKey(prv *ecdsa.PrivateKey) *ecdsa.PublicKey {
	pkBytes := crypto.FromECDSA(prv)
	pk := crypto.ToECDSAPub(pkBytes)
	return pk
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
