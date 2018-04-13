package optimus

import (
	"context"
	"time"
)

type Watcher interface {
	OnRun()
	OnShutdown()
	Execute(ctx context.Context)
}

type managedWatcher struct {
	watcher Watcher
	timeout time.Duration
}

func newManagedWatcher(watcher Watcher, timeout time.Duration) *managedWatcher {
	return &managedWatcher{
		watcher: watcher,
		timeout: timeout,
	}
}

func (m *managedWatcher) Run(ctx context.Context) error {
	m.watcher.OnRun()
	defer m.watcher.OnShutdown()

	timer := time.NewTicker(m.timeout)
	defer timer.Stop()

	m.watcher.Execute(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			m.watcher.Execute(ctx)
		}
	}
}

type reactiveWatcher struct {
	channel <-chan struct{}
	watcher Watcher
}

func newReactiveWatcher(channel <-chan struct{}, watcher Watcher) *reactiveWatcher {
	return &reactiveWatcher{
		channel: channel,
		watcher: watcher,
	}
}

func (m *reactiveWatcher) Run(ctx context.Context) error {
	m.watcher.OnRun()
	defer m.watcher.OnShutdown()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.channel:
			m.watcher.Execute(ctx)
		}
	}
}
