package pubdetector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

// PublicIP detects public IP
func PublicIP() (net.IP, error) {
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
