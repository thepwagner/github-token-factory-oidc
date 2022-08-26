package oidc

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/thepwagner/github-token-action-server/api"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type TracedTokenParser struct {
	tracer  trace.Tracer
	traced  *http.Client
	wrapped api.TokenParser
}

func NewTracedTokenParser(tp trace.TracerProvider, parser api.TokenParser) api.TokenParser {
	return TracedTokenParser{
		tracer: tp.Tracer(""),
		traced: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport, otelhttp.WithTracerProvider(tp)),
		},
		wrapped: parser,
	}
}

var _ api.TokenParser = (*TracedTokenParser)(nil)

func (p TracedTokenParser) Parse(ctx context.Context, tok string) (api.Claims, error) {
	ctx, span := p.tracer.Start(ctx, "TokenParser.Parse")
	defer span.End()

	ctx = oidc.ClientContext(ctx, p.traced)
	claims, err := p.wrapped.Parse(ctx, tok)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return claims, err
}
