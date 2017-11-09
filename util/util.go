package util

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os/user"
	"runtime"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v2"
	"os"
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

// GetPublicIP detects public IP
func GetPublicIP() (net.IP, error) {
	req, err := http.NewRequest("GET", "http://checkip.amazonaws.com/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK response from checkip.amamazonaws.com: %v", resp.Status)
	}

	n := bytes.IndexByte(body, '\n')
	s := string(body[:n])

	pubipadr := net.ParseIP(s)
	if pubipadr == nil {
		return nil, fmt.Errorf("failed to ParseIP from: %s", s)
	}
	return pubipadr, nil
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
		return "", errors.New("Invalid port value")
	}

	return port, nil
}

func GetPlatformName() string {
	return fmt.Sprintf("%s/%s/%s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func PubKeyToString(key ecdsa.PublicKey) string {
	return fmt.Sprintf("%x", crypto.FromECDSAPub(&key))
}

func PubKeyToAddr(key ecdsa.PublicKey) string {
	return crypto.PubkeyToAddress(key).String()
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

// DirectoryExists returns true if given directory exists
func DirectoryExists(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
