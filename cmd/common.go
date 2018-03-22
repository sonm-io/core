package cmd

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitInterrupted() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
