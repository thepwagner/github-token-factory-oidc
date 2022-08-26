package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/thepwagner/github-token-action-server/api"
)

// TokenParser parses tokens from a known issuer
type TokenParser struct {
	verifier *oidc.IDTokenVerifier
}

var _ api.TokenParser = (*TokenParser)(nil)

func NewTokenParser(ctx context.Context, issuer string, opts ...TokenParserOpt) (*TokenParser, error) {
	prov, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("creating provider: %w", err)
	}
	cfg := &oidc.Config{
		SkipClientIDCheck: true,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	verifier := prov.Verifier(cfg)
	return &TokenParser{verifier: verifier}, nil
}

type TokenParserOpt func(*oidc.Config)

func (p *TokenParser) Parse(ctx context.Context, tok string) (api.Claims, error) {
	parsed, err := p.verifier.Verify(ctx, tok)
	if err != nil {
		return nil, fmt.Errorf("verifying token: %w", err)
	}
	var claims api.Claims
	if err := parsed.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parsing claims: %w", err)
	}
	return claims, nil
}
