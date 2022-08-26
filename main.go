package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/thepwagner/github-token-action-server/server"
)

func main() {
	zl := zerolog.New(os.Stderr)
	zl = zl.With().Timestamp().Logger()
	var log logr.Logger = zerologr.New(&zl)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sig
		log.V(1).Info("received signal", "signal", signal)
		cancel()
	}()

	if err := server.Run(ctx, log); err != nil {
		log.Error(err, "failed to run")
		os.Exit(1)
	}
}
