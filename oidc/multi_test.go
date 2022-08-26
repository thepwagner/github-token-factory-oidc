package oidc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/github-token-action-server/oidc"
)

func TestMultiIssuerParser_Actions(t *testing.T) {
	ctx := context.Background()

	t.Run("issuer found", func(t *testing.T) {
		tp, err := oidc.NewMultiIssuerParser(ctx, issuerGitHub)
		require.NoError(t, err)

		_, err = tp.Parse(ctx, ghToken)
		assert.Error(t, err, "token is expired")
	})

	t.Run("issuer not found", func(t *testing.T) {
		tp, err := oidc.NewMultiIssuerParser(ctx, issuerGoogle)
		require.NoError(t, err)
		_, err = tp.Parse(ctx, ghToken)
		assert.Error(t, err, "token is expired")
	})
}
