package util

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"

	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/hashicorp/go-multierror"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

const (
	FormattedBigIntLength = 80
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

// StringToEtherPrice converts input string s to Ethereum's price present as big.Int
// This function expects to receive trimmed input.
func StringToEtherPrice(s string) (*big.Int, error) {
	bigFloat, ok := big.NewFloat(0).SetString(s)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to float value", s)
	}

	if bigFloat.Cmp(big.NewFloat(0)) < 0 {
		return nil, errors.New("value cannot be negative")
	}

	v, _ := big.NewFloat(0).Mul(bigFloat, big.NewFloat(params.Ether)).Int(nil)

	if v.Cmp(big.NewInt(0)) == 0 && bigFloat.Cmp(big.NewFloat(0)) > 0 {
		return nil, errors.New("value is too low")
	}

	return v, nil
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

func IsPrivateIP(ip net.IP) bool {
	return isPrivateIPv4(ip) || isPrivateIPv6(ip)
}

func isPrivateIPv4(ip net.IP) bool {
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	return private24BitBlock.Contains(ip) ||
		private20BitBlock.Contains(ip) ||
		private16BitBlock.Contains(ip) ||
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast()
}

func isPrivateIPv6(ip net.IP) bool {
	_, block, _ := net.ParseCIDR("fc00::/7")

	return block.Contains(ip) || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func StartPrometheus(ctx context.Context, listenAddr string) {
	log.GetLogger(ctx).Info(
		"starting metrics server", zap.String("metrics_addr", listenAddr))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(listenAddr, nil)
}

func BigIntToPaddedString(x *big.Int) string {
	raw := x.String()
	paddingSize := FormattedBigIntLength - len(raw)
	padding := make([]rune, paddingSize)
	for idx := range padding {
		padding[idx] = '0'
	}

	return string(padding) + raw
}

func MultierrFormat() multierror.ErrorFormatFunc {
	return func(errs []error) string {
		var s []string
		for _, e := range errs {
			s = append(s, e.Error())
		}

		return strings.Join(s, ", ")
	}
}

func HexToAddress(hex string) (common.Address, error) {
	if !common.IsHexAddress(hex) {
		return common.Address{}, fmt.Errorf("invalid address %s specified for parsing", hex)
	}
	return common.HexToAddress(hex), nil
}

func LaconicError(err error) zapcore.Field {
	return zap.String("error", err.Error())
}
