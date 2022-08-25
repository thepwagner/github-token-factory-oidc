package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/thepwagner/github-token-action-server/api"
	"github.com/thepwagner/github-token-action-server/github"
)

func run(ctx context.Context, log logr.Logger) error {
	parser, err := github.NewOIDCParser(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OIDC parser: %w", err)
	}

	authz := api.TokenCheckYOLO

	// issuer, err := github.NewTokenIssuer(log)
	// if err != nil {
	// 	return fmt.Errorf("failed to create token issuer: %w", err)
	// }
	var issuer api.TokenIssuer

	srv := http.Server{
		Addr:    ":8080",
		Handler: api.NewHandler(log, parser.Parse, authz, issuer),
	}

	log.Info("starting server", "addr", srv.Addr)
	errChan := make(chan error, 1)
	go func() { errChan <- srv.ListenAndServe() }()
	select {
	case <-errChan:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		log.V(1).Info("server shutdown complete")
	}

	return nil
}

func main() {
	zl := zerolog.New(zerolog.NewConsoleWriter())
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

	if err := run(ctx, log); err != nil {
		log.Error(err, "failed to run")
		os.Exit(1)
	}
}
