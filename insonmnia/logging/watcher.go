package logging

import (
	"fmt"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Watcher struct {
	zapcore.Core

	mu        sync.RWMutex
	observers map[string]chan<- string
}

func (m *Watcher) With(fields []zapcore.Field) zapcore.Core {
	m.Core.With(fields)
	return m
}

func (m *Watcher) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checkedEntry.AddCore(entry, m)
}

func (m *Watcher) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	buf, err := encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, observer := range m.observers {
		observer <- buf.String()
	}

	return m.Core.Write(entry, fields)
}

func NewWatcher() *Watcher {
	return &Watcher{
		observers: map[string]chan<- string{},
	}
}

func (m *Watcher) OnLog(entry zapcore.Entry) error {
	message := fmt.Sprintf("%s\t%s\t%s\t%s", entry.Time.Format(time.RFC3339Nano), entry.Level.CapitalString(), entry.Caller.TrimmedPath(), entry.Message)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, observer := range m.observers {
		observer <- message
	}

	return nil
}

func (m *Watcher) Subscribe(tx chan<- string) string {
	id := uuid.New()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.observers[id] = tx

	return id
}

func (m *Watcher) Unsubscribe(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.observers, id)
}
