package github

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/thepwagner/github-token-action-server/api"
)

type OIDCParser struct {
	verifier *oidc.IDTokenVerifier
}

func NewOIDCParser(ctx context.Context, opts ...TokenParserOpt) (*OIDCParser, error) {
	prov, err := oidc.NewProvider(ctx, "https://token.actions.githubusercontent.com")
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
	return &OIDCParser{verifier: verifier}, nil
}

type TokenParserOpt func(*oidc.Config)

func (p *OIDCParser) Parse(ctx context.Context, tok string) (*api.WorkflowID, error) {
	parsed, err := p.verifier.Verify(ctx, tok)
	if err != nil {
		return nil, fmt.Errorf("verifying token: %w", err)
	}
	var id api.WorkflowID
	if err := parsed.Claims(&id); err != nil {
		return nil, fmt.Errorf("parsing claims: %w", err)
	}
	return &id, nil
}
