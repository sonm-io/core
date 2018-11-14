package xnet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// LookupLoopbackIP looks up loopback interfaces on the host using the local
// resolver.
// It returns a slice of that host's IPv6 and IPv4 addresses.
func LookupLoopbackIP() ([]net.IP, error) {
	links, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// Distinguishing between IPv6 and IPv4 addresses if required to stick to
	// the RFC 6724, which requires IPv6 to be before IPv4.
	var loV6Addrs []net.IP
	var loV4Addrs []net.IP

	for _, link := range links {
		// Ignore down and non-loopback links. However, link-local links
		// like under the "fe80::/10" subnet for example still need to be
		// filtered out later.
		if link.Flags&(net.FlagUp|net.FlagLoopback) != (net.FlagUp | net.FlagLoopback) {
			continue
		}

		addrs, err := link.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			if ip.IsLoopback() {
				if ip := ip.To16(); ip != nil {
					loV6Addrs = append(loV6Addrs, ip)
					continue
				}

				if ip := ip.To4(); ip != nil {
					loV4Addrs = append(loV4Addrs, ip)
					continue
				}
			}
		}
	}

	return append(loV6Addrs, loV4Addrs...), nil
}

// ExternalPublicIPResolver is a helper struct that allows to resolve caller's
// public IP address.
type ExternalPublicIPResolver struct {
	url string
	// Cached IP.
	ip            net.IP
	cacheTime     time.Time
	cacheDuration time.Duration
}

// NewExternalPublicIPResolver constructs a new external public IP resolver.
//
// An optional "url" argument specifies the server URL, which can give the
// caller's public IP in a body as a string.
func NewExternalPublicIPResolver(url string) *ExternalPublicIPResolver {
	if url == "" {
		url = "http://checkip.amazonaws.com/"
	}

	return &ExternalPublicIPResolver{
		url:           url,
		cacheDuration: 10 * time.Minute,
	}
}

func (m *ExternalPublicIPResolver) PublicIP() (net.IP, error) {
	if m.needRefresh() {
		if err := m.refresh(); err != nil {
			return nil, err
		}
	}

	return m.ip, nil
}

func (m *ExternalPublicIPResolver) needRefresh() bool {
	return m.ip == nil || time.Now().Sub(m.cacheTime) > m.cacheDuration
}

func (m *ExternalPublicIPResolver) refresh() error {
	ip, err := m.resolve()
	if err != nil {
		return err
	}

	m.ip = ip
	m.cacheTime = time.Now()
	return nil
}

func (m *ExternalPublicIPResolver) resolve() (net.IP, error) {
	resp, err := http.Get(m.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resolve external IP address: %s", resp.Status)
	}

	addr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return net.ParseIP(string(bytes.TrimSpace(addr))), nil
}
