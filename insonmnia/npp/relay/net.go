package relay

import (
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	minSleepInterval = 5 * time.Millisecond
	maxSleepInterval = 1 * time.Second
	sleepMultiplier  = 2
)

type BackPressureListener struct {
	net.Listener

	Log *zap.Logger
}

func (m *BackPressureListener) Accept() (net.Conn, error) {
	interval := minSleepInterval

	for {
		conn, err := m.Listener.Accept()
		if err == nil {
			return conn, nil
		}

		if netError, ok := err.(net.Error); ok && netError.Temporary() {
			if max := maxSleepInterval; interval > max {
				interval = max
			}

			m.Log.Warn("failed to accept connection", zap.Error(netError), zap.Duration("sleep", interval))
			time.Sleep(interval)

			interval *= sleepMultiplier
		} else {
			return nil, err
		}
	}
}
