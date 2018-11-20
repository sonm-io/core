package util

import (
	"time"
)

type ImmediateTicker struct {
	ticker   *time.Ticker
	C        <-chan time.Time // The channel on which the ticks are delivered.
	cancelCh chan struct{}
}

func NewImmediateTicker(d time.Duration) *ImmediateTicker {
	ch := make(chan time.Time, 1)
	ticker := &ImmediateTicker{
		ticker:   time.NewTicker(d),
		C:        ch,
		cancelCh: make(chan struct{}),
	}

	go func() {
		ch <- time.Now()
		for {
			select {
			case timePoint := <-ticker.ticker.C:
				select {
				case ch <- timePoint:
				default:
				}
			case <-ticker.cancelCh:
				return
			}

		}
	}()

	return ticker
}

func (t *ImmediateTicker) Stop() {
	t.ticker.Stop()
	t.cancelCh <- struct{}{}
}
