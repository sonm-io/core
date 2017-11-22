package miner

import (
	"net"
	"sort"
	"sync"
	"time"
)

// BackoffTimer implementation
type BackoffTimer struct {
	sleep    time.Duration
	maxsleep time.Duration
	t        *time.Timer
}

// NewBackoffTimer implementations one direction backoff policy
func NewBackoffTimer(sleep, maxsleep time.Duration) *BackoffTimer {
	bt := &BackoffTimer{
		sleep:    sleep,
		maxsleep: maxsleep,
		t:        time.NewTimer(0),
	}
	return bt
}

// C resets Timer and returns Timer.C
func (b *BackoffTimer) C() <-chan time.Time {
	b.sleep *= 2
	if b.sleep > b.maxsleep {
		b.sleep = b.maxsleep
	}

	if !b.t.Stop() {
		<-b.t.C
	}
	b.t.Reset(b.sleep)
	return b.t.C
}

// Stop frees the Timer
func (b *BackoffTimer) Stop() bool {
	return b.t.Stop()
}

var stringArrayPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 10)
	},
}

func SortedIPs(ips []string) []string {
	var sorted sortableIPs
	for _, strIP := range ips {
		if ip := net.ParseIP(strIP); ip != nil {
			sorted = append(sorted, ip)
		}
	}
	sort.Sort(sorted)

	out := make([]string, len(sorted))
	for idx, ip := range sorted {
		out[idx] = ip.String()
	}

	return out
}

// Sorting is implemented as follows: first come all public IPs (IPv6 before IPv4), then
// all private IPs (IPv6 before IPv4).
type sortableIPs []net.IP

func (s sortableIPs) Len() int      { return len(s) }
func (s sortableIPs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableIPs) Less(i, j int) bool {
	iIsPrivate, jIsPrivate := isPrivate(s[i]), isPrivate(s[j])
	if iIsPrivate && !jIsPrivate {
		return false
	}

	if jIsPrivate && !iIsPrivate {
		return true
	}

	// Both are private, check for family.
	iIsIPv4, jIsIPv4 := isIPv4(s[i]), isIPv4(s[j])
	if iIsIPv4 && !jIsIPv4 {
		return false
	}

	return true
}

func isPrivate(ip net.IP) bool {
	return isPrivateIPv4(ip) || isPrivateIPv6(ip)
}

func isPrivateIPv4(ip net.IP) bool {
	private := false
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	private = private24BitBlock.Contains(ip) || private20BitBlock.Contains(ip) || private16BitBlock.Contains(ip)

	return private
}

func isPrivateIPv6(ip net.IP) bool {
	_, block, _ := net.ParseCIDR("fc00::/7")

	return block.Contains(ip)
}

func isIPv4(ip net.IP) bool {
	return ip.To4() != nil
}
