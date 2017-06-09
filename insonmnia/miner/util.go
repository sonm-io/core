package miner

import (
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
