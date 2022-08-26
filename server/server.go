package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/thepwagner/github-token-action-server/api"
	"github.com/thepwagner/github-token-action-server/github"
	"github.com/thepwagner/github-token-action-server/oidc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func Run(ctx context.Context, log logr.Logger) error {
	cfg, err := NewConfig()
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}
	tp, err := newTracerProvider(cfg)
	if err != nil {
		return fmt.Errorf("building tracer: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Error(err, "failed to shutdown tracer")
		}
		log.V(2).Info("tracer shutdown complete")
	}()

	parser, err := oidc.NewParser(ctx, cfg.Issuers...)
	if err != nil {
		return fmt.Errorf("failed to create OIDC parser: %w", err)
	}

	authz := api.TokenCheckYOLO

	issuer := github.NewIssuer(log, tp, cfg.GitHub)

	handler := api.NewHandler(log, tp.Tracer(""), parser, authz, issuer.IssueToken)
	traced := otelhttp.NewHandler(handler, "ServeHTTP", otelhttp.WithTracerProvider(tp))
	return runServer(ctx, log, cfg.Addr, traced)
}

func newTracerProvider(cfg *Config) (*sdktrace.TracerProvider, error) {
	tpOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tft"),
		)),
	}
	if cfg.JaegerEndpoint != "" {
		jaegerOut, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerEndpoint)))
		if err != nil {
			return nil, err
		}
		tpOptions = append(tpOptions, sdktrace.WithBatcher(jaegerOut))
	}

	return sdktrace.NewTracerProvider(tpOptions...), nil
}

func runServer(ctx context.Context, log logr.Logger, addr string, handler http.Handler) error {
	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	log.Info("starting server", "addr", srv.Addr)
	errChan := make(chan error, 1)
	go func() { errChan <- srv.ListenAndServe() }()
	select {
	case err := <-errChan:
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
