package npp

import (
	"net"
)

type connResult struct {
	conn net.Conn
	err  error
}

func newConnResult(conn net.Conn, err error) connResult {
	return connResult{conn, err}
}

func newConnResultOk(conn net.Conn) connResult {
	return newConnResult(conn, nil)
}

func newConnResultErr(err error) connResult {
	return newConnResult(nil, err)
}

func (m *connResult) RemoteAddr() net.Addr {
	if m == nil || m.conn == nil {
		return nil
	}
	return m.conn.RemoteAddr()
}

func (m *connResult) Close() error {
	if m == nil || m.conn == nil {
		return nil
	}
	return m.conn.Close()
}

func (m *connResult) Error() error {
	return m.err
}

func (m *connResult) IsRendezvousError() bool {
	if m.err == nil {
		return false
	}

	_, ok := m.err.(*rendezvousError)
	return ok
}

func (m *connResult) Unwrap() (net.Conn, error) {
	return m.conn, m.err
}

func (m *connResult) UnwrapWithSource(source connSource) (net.Conn, connSource, error) {
	return m.conn, source, m.err
}
