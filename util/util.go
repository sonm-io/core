package util

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"os/user"
	"path"
	"runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	FormattedBigIntLength = 80
	HomeConfigDir         = ".sonm"
	DefaultKeystorePath   = "keystore"
)

// GetUserHomeDir returns home for current user.
// e.g: /home/user
func GetUserHomeDir() (homeDir string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

// GetDefaultConfigDir returns default user's config dir.
// e.g: /home/user/.sonm
func GetDefaultConfigDir() (string, error) {
	home, err := GetUserHomeDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(home, HomeConfigDir)
	return dir, nil
}

// GetDefaultKeyStoreDir returns default keystore dir for current user.
// e.g: /home/user/.sonm/keystore
func GetDefaultKeyStoreDir() (string, error) {
	cfgDir, err := GetDefaultConfigDir()
	if err != nil {
		return "", err
	}

	keyDir := path.Join(cfgDir, DefaultKeystorePath)
	return keyDir, nil
}

func GetPlatformName() string {
	return fmt.Sprintf("%s/%s/%s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func FileExists(p string) error {
	f, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", p)
		}
		return fmt.Errorf("failed to access %s: %s", p, err)
	}
	if f.IsDir() {
		return fmt.Errorf("%s is directory", p)
	}
	return nil
}

// DirectoryExists returns true if the given directory exists
func DirectoryExists(p string) error {
	f, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", p)
		}
		return fmt.Errorf("failed to access %s: %s", p, err)
	}

	if !f.IsDir() {
		return fmt.Errorf("%s is not a directory", p)
	}
	return nil
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

func HexToAddress(hex string) (common.Address, error) {
	if !common.IsHexAddress(hex) {
		return common.Address{}, fmt.Errorf("invalid address %s specified for parsing", hex)
	}
	return common.HexToAddress(hex), nil
}
