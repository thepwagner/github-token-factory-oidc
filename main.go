package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/thepwagner/github-token-action-server/api"
	"github.com/thepwagner/github-token-action-server/github"
)

func run(log logr.Logger) error {
	ctx := context.Background()
	parser, err := github.NewOIDCParser(ctx)
	if err != nil {
		return fmt.Errorf("failed to create OIDC parser: %w", err)
	}

	authz := api.TokenCheckYOLO

	issuer, err := github.NewTokenIssuer(log)
	if err != nil {
		return fmt.Errorf("failed to create token issuer: %w", err)
	}

	srv := http.Server{
		Addr:    ":8080",
		Handler: api.NewHandler(log, parser, authz, issuer),
	}
	log.Info("starting server", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func main() {
	zl := zerolog.New(os.Stderr)
	zl = zl.With().Timestamp().Logger()
	var log logr.Logger = zerologr.New(&zl)

	if err := run(log); err != nil {
		log.Error(err, "failed to run")
		os.Exit(1)
	}
}
