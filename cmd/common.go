package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WaitInterrupted(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
	case <-ctx.Done():
	}
}
