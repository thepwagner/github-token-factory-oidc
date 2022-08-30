package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/github-token-factory-oidc/server"
)

func TestNewConfig(t *testing.T) {
	c, err := server.NewConfig()
	require.NoError(t, err)
	assert.Equal(t, ".github", c.Checker.Rego.OwnerRepo)
	assert.Equal(t, false, c.Checker.Rego.FromRepos)
}
