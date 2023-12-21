package github_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/github-token-factory-oidc/api"
	"github.com/thepwagner/github-token-factory-oidc/github"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestIssuer(t *testing.T) {
	t.Parallel()
	t.Skip("uses a private key and live GitHub")

	configs := map[string]github.Config{
		"*": {
			AppID:          89357,
			PrivateKeyPath: "/Users/pwagner/Downloads/actions-secret-garden.2022-08-25.private-key.pem",
		},
	}

	iss := github.NewIssuer(slog.Default(), noop.NewTracerProvider().Tracer(""), github.NewClients(http.DefaultTransport, configs))

	tok, err := iss.IssueToken(context.Background(), &api.TokenRequest{
		Repositories: []string{"thepwagner-org/debian-bullseye"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tok)
}
