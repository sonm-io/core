package rendezvous

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type resolver interface {
	// PublicIP resolves this's host public IP.
	PublicIP() (net.IP, error)
}

type externalResolver struct {
	url string
	// Cached IP.
	ip            net.IP
	cacheTime     time.Time
	cacheDuration time.Duration
}

func newExternalResolver(url string) resolver {
	if url == "" {
		url = "http://checkip.amazonaws.com/"
	}

	return &externalResolver{
		url:           url,
		cacheDuration: 10 * time.Minute,
	}
}

func (m *externalResolver) PublicIP() (net.IP, error) {
	if m.needRefresh() {
		if err := m.refresh(); err != nil {
			return nil, err
		}
	}

	return m.ip, nil
}

func (m *externalResolver) needRefresh() bool {
	return m.ip == nil || time.Now().Sub(m.cacheTime) > m.cacheDuration
}

func (m *externalResolver) refresh() error {
	ip, err := m.resolve()
	if err != nil {
		return err
	}

	m.ip = ip
	m.cacheTime = time.Now()
	return nil
}

func (m *externalResolver) resolve() (net.IP, error) {
	resp, err := http.Get(m.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resolve external IP address: %s", resp.Status)
	}

	addr, err := ioutil.ReadAll(resp.Body)
	return net.ParseIP(string(addr)), nil
}
