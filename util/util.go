package util

import (
	"fmt"
	"net"

	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os/user"
	"strconv"
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
