package logging

import (
	"sync"

	"github.com/pborman/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type watcher struct {
	mu        sync.RWMutex
	observers map[string]chan<- string
}

func newWatcher() *watcher {
	return &watcher{
		observers: map[string]chan<- string{},
	}
}

func (m *watcher) Notify(message string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, observer := range m.observers {
		select {
		case observer <- message:
		default:
			// Seems like the receiver side is blocked. If the channel capacity
			// is large enough it may still receive missing messages. Otherwise
			// trying to block above results in freezing the entire worker,
			// which is not what we want.
		}
	}
}

func (m *watcher) Subscribe(tx chan<- string) string {
	id := uuid.New()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.observers[id] = tx

	return id
}

func (m *watcher) Unsubscribe(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.observers, id)
}

type WatcherCore struct {
	zapcore.Core

	encoder zapcore.Encoder
	watcher *watcher
}

func NewWatcherCore() *WatcherCore {
	return &WatcherCore{
		encoder: zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		watcher: newWatcher(),
	}
}

func (m *WatcherCore) With(fields []zapcore.Field) zapcore.Core {
	encoder := m.encoder.Clone()
	for _, field := range fields {
		field.AddTo(encoder)
	}

	return &WatcherCore{
		Core:    m.Core.With(fields),
		encoder: encoder,
		watcher: m.watcher,
	}
}

func (m *WatcherCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checkedEntry.AddCore(entry, m)
}

func (m *WatcherCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	buf, err := m.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	m.watcher.Notify(buf.String())

	return m.Core.Write(entry, fields)
}

func (m *WatcherCore) Subscribe(tx chan<- string) string {
	return m.watcher.Subscribe(tx)
}

func (m *WatcherCore) Unsubscribe(id string) {
	m.watcher.Unsubscribe(id)
}
