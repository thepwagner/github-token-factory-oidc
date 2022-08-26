package oidc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/thepwagner/github-token-factory-oidc/api"
)

// NewParser returns an appropriate parser
func NewParser(ctx context.Context, issuers ...string) (api.TokenParser, error) {
	switch len(issuers) {
	case 0:
		return nil, fmt.Errorf("no issuers")
	case 1:
		return NewTokenParser(ctx, issuers[0])
	default:
		return NewMultiIssuerParser(ctx, issuers...)
	}
}

// MultiIssuerParser supports multiple issuers
type MultiIssuerParser struct {
	issuers map[string]api.TokenParser
}

var _ api.TokenParser = (*MultiIssuerParser)(nil)

func NewMultiIssuerParser(ctx context.Context, issuers ...string) (*MultiIssuerParser, error) {
	parsers := make(map[string]api.TokenParser, len(issuers))
	for _, issuer := range issuers {
		parser, err := NewTokenParser(ctx, issuer)
		if err != nil {
			return nil, fmt.Errorf("creating parser for issuer %q: %w", issuer, err)
		}
		parsers[issuer] = parser
	}
	return &MultiIssuerParser{issuers: parsers}, nil
}

func (p *MultiIssuerParser) Parse(ctx context.Context, tok string) (api.Claims, error) {
	// The delegated parser will validate, we can just hack the issuer out of the JWT:
	toks := strings.Split(tok, ".")
	if len(toks) != 3 {
		return nil, fmt.Errorf("invalid token")
	}
	var payload idToken
	decoded, err := base64.RawURLEncoding.DecodeString(toks[1])
	if err != nil {
		return nil, fmt.Errorf("base64 decoding JWT payload: %w", err)
	}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("unmarshaling JWT payload: %w", err)
	}

	parser, ok := p.issuers[payload.Issuer]
	if !ok {
		return nil, fmt.Errorf("no parser for issuer %q", payload.Issuer)
	}
	return parser.Parse(ctx, tok)
}

type idToken struct {
	Issuer string `json:"iss"`
}
