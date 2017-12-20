package util

import "time"

type ImmediateTicker struct {
	ticker   *time.Ticker
	C        <-chan time.Time // The channel on which the ticks are delivered.
	cancelCh chan struct{}
}

func NewImmediateTicker(d time.Duration) *ImmediateTicker {
	ch := make(chan time.Time, 1)
	it := ImmediateTicker{
		ticker:   time.NewTicker(d),
		C:        ch,
		cancelCh: make(chan struct{}),
	}
	go func() {
		ch <- time.Now()
		for {
			select {
			case time := <-it.ticker.C:
				select {
				case ch <- time:
				default:
				}
			case <-it.cancelCh:
				return
			}

		}
	}()
	return &it
}

func (t *ImmediateTicker) Stop() {
	t.ticker.Stop()
	t.cancelCh <- struct{}{}
}
