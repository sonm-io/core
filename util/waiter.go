package util

import "sync"

type Waiter struct {
	wg sync.WaitGroup
}

func (w *Waiter) Run(routine func()) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		routine()
	}()
}

func (w *Waiter) Wait() {
	w.wg.Wait()
}
