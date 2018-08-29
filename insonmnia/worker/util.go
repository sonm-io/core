package worker

import (
	"io"
	"sync"
	"time"

	"github.com/sonm-io/core/proto"
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
