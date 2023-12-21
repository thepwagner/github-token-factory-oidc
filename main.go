package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	"github.com/thepwagner/github-token-factory-oidc/server"
)

func main() {
	log := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sig
		log.Debug("received signal", "signal", signal)
		cancel()
	}()

	if err := server.Run(ctx, log); err != nil {
		log.Error("failed to run", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
