package checker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/open-policy-agent/opa/rego"
	"github.com/thepwagner/github-token-factory-oidc/api"
)

// Rego is an api.TokenChecker that evaluates a Rego policy.
type Rego struct {
	log   logr.Logger
	query rego.PreparedEvalQuery
}

func NewRego(ctx context.Context, log logr.Logger, policy string) (*Rego, error) {
	query, err := rego.New(
		rego.Query("data.tokens.allow"),
		rego.Module("tokens.rego", policy),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("preparing query: %w", err)
	}
	return &Rego{
		log:   log,
		query: query,
	}, nil
}

var _ api.TokenChecker = (*Rego)(nil)

type regoInput struct {
	Claims       api.Claims        `json:"claims"`
	Repositories []string          `json:"repositories"`
	Permissions  map[string]string `json:"permissions"`
}

func (r Rego) Check(ctx context.Context, claims api.Claims, req *api.TokenRequest) (bool, error) {
	ri := regoInput{
		Claims:       claims,
		Repositories: req.Repositories,
		Permissions:  req.Permissions,
	}
	riJSON, _ := json.Marshal(ri)
	r.log.Info("evaluating policy", "input", string(riJSON))

	rs, err := r.query.Eval(ctx, rego.EvalInput(ri))
	if err != nil {
		return false, fmt.Errorf("evaluating query: %w", err)
	}
	return rs.Allowed(), nil
}
