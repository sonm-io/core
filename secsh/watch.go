package secsh

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
)

// WatchDir blocks until either the directory on the specified path are made
// exist or the context is canceled.
func WatchDir(ctx context.Context, path string, interval time.Duration, log *zap.SugaredLogger) error {
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if _, err := os.Stat(path); err == nil {
				return nil
			}

			log.Infof("path %s not exists, waiting", path)
		}
	}
}
