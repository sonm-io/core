package util

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/user"
	"runtime"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v2"
)

// GetLocalIP find local non-loopback ip addr
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

func GetUserHomeDir() (homeDir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func ParseEndpointPort(s string) (string, error) {
	_, port, err := net.SplitHostPort(s)
	if err != nil {
		return "", err
	}

	intPort, err := strconv.Atoi(port)
	if err != nil {
		return "", err
	}

	if intPort < 1 || intPort > 65535 {
		return "", errors.New("invalid port value")
	}

	return port, nil
}

func GetPlatformName() string {
	return fmt.Sprintf("%s/%s/%s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func PubKeyToString(key ecdsa.PublicKey) string {
	return fmt.Sprintf("%x", crypto.FromECDSAPub(&key))
}

func PubKeyToAddr(key ecdsa.PublicKey) common.Address {
	return crypto.PubkeyToAddress(key)
}

func LoadYamlFile(from string, to interface{}) error {
	buf, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, to)
	if err != nil {
		return err
	}

	return nil
}

// DirectoryExists returns true if the given directory exists
func DirectoryExists(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// ParseBigInt parses the given string and converts it to *big.Int
func ParseBigInt(s string) (*big.Int, error) {
	n := new(big.Int)
	n, ok := n.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to big.Int", s)
	}

	return n, nil
}

func GetAvailableIPs() (availableIPs []net.IP, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.IsGlobalUnicast() {
				availableIPs = append(availableIPs, ip)
			}
		}
	}
	if len(availableIPs) == 0 {
		return nil, errors.New("could not determine a single unicast addr, check networking")
	}

	return availableIPs, nil
}

// EqualAddresses checks if provided Addresses are equal
func EqualAddresses(a, b common.Address) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
