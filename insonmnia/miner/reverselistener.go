package miner

import (
	"errors"
	"net"
	"sync/atomic"
)

var (
	// ErrListenerClosed is returned to clients and servers is the listener closed
	ErrListenerClosed = errors.New("inmemory listener closed")
)

type reverseListener struct {
	queue   chan net.Conn
	onClose chan struct{}

	closed uint64
}

var _ net.Listener = &reverseListener{}

// NewReverseListener returns inmemory Listener
func newReverseListener(backlog int) *reverseListener {
	return &reverseListener{
		queue:   make(chan net.Conn, backlog),
		onClose: make(chan struct{}),
		closed:  uint64(0),
	}
}

func (rl *reverseListener) Accept() (net.Conn, error) {
	select {
	case <-rl.onClose:
		return nil, ErrListenerClosed
	case c := <-rl.queue:
		return c, nil
	}
}

func (rl *reverseListener) Addr() net.Addr {
	return &net.UnixAddr{Name: "", Net: "tcp"}
}

func (rl *reverseListener) Close() error {
	if atomic.CompareAndSwapUint64(&rl.closed, 0, 1) {
		close(rl.onClose)
	}
	// TODO: dispose queued connections
	return nil
}

func (rl *reverseListener) enqueue(conn net.Conn) error {
	select {
	case rl.queue <- conn:
		return nil
	case <-rl.onClose:
		conn.Close()
		return ErrListenerClosed
	}
}
