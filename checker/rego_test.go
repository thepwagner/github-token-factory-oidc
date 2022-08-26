package checker_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/github-token-factory-oidc/api"
	"github.com/thepwagner/github-token-factory-oidc/checker"
)

func TestRego(t *testing.T) {
	cases := map[string]struct {
		policy        string
		claims        api.Claims
		expectedAllow []*api.TokenRequest
		expectedDeny  []*api.TokenRequest
	}{
		"deny all": {
			policy: `
				package tokens
				allow = false
			`,
			claims: actionClaims,
			expectedDeny: []*api.TokenRequest{
				readContents, writeContents, writeOrgProjects,
			},
		},
		"allow all": {
			policy: `
				package tokens
				allow = true
			`,
			claims: actionClaims,
			expectedAllow: []*api.TokenRequest{
				readContents, writeContents, writeOrgProjects,
			},
		},
		"allow workflows from private repositories": {
			policy: `
				package tokens
				default allow = false
				allow = true {
					input.claims.repository_owner == "thepwagner"
					input.claims.repository_visibility != "public"
					input.permissions.contents != "write"
				}
			`,
			claims: actionClaims,
			expectedAllow: []*api.TokenRequest{
				readContents,
			},
			expectedDeny: []*api.TokenRequest{
				writeContents, writeOrgProjects,
			},
		},
		"allow workflows with specific permissions": {
			policy: `
				package tokens
				default allow = false
				allow = true {
					input.claims.repository_owner == "thepwagner"
					input.claims.repository_visibility != "public"
					input.permissions == {
						"organization_projects": "write"
					}
				}
			`,
			claims: actionClaims,
			expectedAllow: []*api.TokenRequest{
				writeOrgProjects,
			},
			expectedDeny: []*api.TokenRequest{
				writeContents,
				readContents,
			},
		},
	}

	ctx := context.Background()
	for label, tc := range cases {
		t.Run(label, func(t *testing.T) {
			r, err := checker.NewRego(ctx, logr.Discard(), tc.policy)
			require.NoError(t, err)

			for _, allow := range tc.expectedAllow {
				ok, err := r.Check(ctx, tc.claims, allow)
				require.NoError(t, err)
				assert.True(t, ok)
			}
			for _, deny := range tc.expectedDeny {
				ok, err := r.Check(ctx, tc.claims, deny)
				require.NoError(t, err)
				assert.False(t, ok)
			}
		})
	}
}

var (
	actionClaims = api.Claims{
		"sub":                   "repo:thepwagner/github-token-action:ref:refs/heads/main",
		"ref":                   "refs/heads/main",
		"sha":                   "3b913b904ae6ce79be995469d302c4f82cabb712",
		"repository":            "thepwagner/github-token-action",
		"repository_owner":      "thepwagner",
		"repository_owner_id":   "1559510",
		"run_id":                "2676512750",
		"run_number":            "5",
		"run_attempt":           "1",
		"repository_visibility": "private",
		"repository_id":         "514238596",
		"actor_id":              "1559510",
		"actor":                 "thepwagner",
		"workflow":              "Hack",
		"head_ref":              "",
		"base_ref":              "",
		"event_name":            "push",
		"ref_type":              "branch",
		"job_workflow_ref":      "thepwagner/github-token-action/.github/workflows/hack.yml@refs/heads/main",
		"iss":                   "https://token.actions.githubusercontent.com",
	}
	readContents = &api.TokenRequest{
		Permissions: map[string]string{
			"contents": "read",
		},
	}
	writeContents = &api.TokenRequest{
		Permissions: map[string]string{
			"contents": "write",
		},
	}
	writeOrgProjects = &api.TokenRequest{
		Permissions: map[string]string{
			"organization_projects": "write",
		},
	}
)
