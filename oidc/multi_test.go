package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/github-token-factory-oidc/oidc"
)

func TestMultiIssuerParser_Actions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("issuer found", func(t *testing.T) {
		t.Parallel()
		tp, err := oidc.NewMultiIssuerParser(ctx, issuerGitHub)
		require.NoError(t, err)

		_, err = tp.Parse(ctx, ghToken)
		assert.Error(t, err, "token is expired")
	})

	t.Run("issuer not found", func(t *testing.T) {
		t.Parallel()
		tp, err := oidc.NewMultiIssuerParser(ctx, issuerGoogle)
		require.NoError(t, err)
		_, err = tp.Parse(ctx, ghToken)
		assert.Error(t, err, "token is expired")
	})
}
