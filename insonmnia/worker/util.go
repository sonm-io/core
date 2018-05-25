package worker

import (
	"io"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/netutil"
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
	iIsPrivate, jIsPrivate := netutil.IsPrivateIP(s[i]), netutil.IsPrivateIP(s[j])
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

func isIPv4(ip net.IP) bool {
	return ip.To4() != nil
}

type chunkReader struct {
	stream sonm.Worker_PushTaskServer
	buf    []byte
}

func newChunkReader(stream sonm.Worker_PushTaskServer) io.Reader {
	return &chunkReader{stream: stream, buf: nil}
}

func (r *chunkReader) Read(p []byte) (n int, err error) {
	// Pull the next chunk when we've completely consumed the current one.
	if len(r.buf) == 0 {
		chunk, err := r.stream.Recv()
		if err != nil {
			if err != io.EOF {
				return 0, err
			}
		}

		if chunk == nil {
			return 0, io.EOF
		}

		r.buf = chunk.Chunk
	}

	size := copy(p, r.buf)

	r.buf = r.buf[size:]

	if err := r.stream.Send(&sonm.Progress{Size: int64(size)}); err != nil {
		return 0, err
	}

	return size, nil
}
