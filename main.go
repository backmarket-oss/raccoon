package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/backmarket-oss/raccoon/cmd"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	ctx := context.Background()
	ctxWithCancel, cancelFct := context.WithCancel(ctx)

	// We want to gracefully shutdown raccoon
	// On signals we propagate context cancelled to the stack
	go func() {
		<-signalChan
		cancelFct()
	}()

	if err := cmd.Execute(ctxWithCancel); err != nil {
		os.Exit(1)
	}
}
