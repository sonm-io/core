package xnet

import (
	"context"
	"fmt"
	"net"
)

type ListenerExt interface {
	net.Listener
	// AcceptContext extends the "Accept" method by allowing to specify the
	// context, which interrupts accepting connections when canceled.
	AcceptContext(ctx context.Context) (net.Conn, error)
}

type connOrError struct {
	conn net.Conn
	err  error
}

func newConnOrError(conn net.Conn, err error) connOrError {
	return connOrError{
		conn: conn,
		err:  err,
	}
}

func (m *connOrError) Unwrap() (net.Conn, error) {
	return m.conn, m.err
}

type ctxListenerWrapper struct {
	net.Listener
	pending <-chan connOrError
}

func WithContext(listener net.Listener) ListenerExt {
	return newCtxListenerWrapper(listener)
}

func newCtxListenerWrapper(listener net.Listener) *ctxListenerWrapper {
	txrx := make(chan connOrError, 0)

	m := &ctxListenerWrapper{
		Listener: listener,
		pending:  txrx,
	}

	go m.run(txrx)

	return m
}

func (m *ctxListenerWrapper) AcceptContext(ctx context.Context) (net.Conn, error) {
	select {
	case message, ok := <-m.pending:
		if ok {
			return message.Unwrap()
		}

		return nil, fmt.Errorf("use of closed network connection")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *ctxListenerWrapper) Close() error {
	// Pending "Accept" should terminate after this call.
	err := m.Listener.Close()

	// Exhaust the pending channel to prevent fd and goroutine leakage.
	for message := range m.pending {
		if message.conn != nil {
			message.conn.Close()
		}
	}

	return err
}

func (m *ctxListenerWrapper) run(tx chan<- connOrError) {
	// Explicitly close the channel when finished, otherwise draining
	// in "Close" will hang forever.
	defer close(tx)

	for {
		conn, err := m.Listener.Accept()
		tx <- newConnOrError(conn, err)

		// Here is the termination condition. Unrecoverable network errors
		// should
		if ne, ok := err.(net.Error); ok {
			if !ne.Temporary() {
				break
			}
		}
	}
}
