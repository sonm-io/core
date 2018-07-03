package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

func WaitInterrupted(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case v := <-sigChan:
		return errors.New(v.String())
	case <-ctx.Done():
		return ctx.Err()
	}
}
