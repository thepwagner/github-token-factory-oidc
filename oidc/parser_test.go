package oidc_test

import (
	"context"
	"testing"
	"time"

	coreoidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/github-token-action-server/oidc"
)

const (
	// This token is expired
	ghToken      = `eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6ImVCWl9jbjNzWFlBZDBjaDRUSEJLSElnT3dPRSIsImtpZCI6Ijc4MTY3RjcyN0RFQzVEODAxREQxQzg3ODRDNzA0QTFDODgwRUMwRTEifQ.eyJqdGkiOiI0YjMyZWExMi1iZDYwLTQ0ODEtOTZhOC0wZDU2ODE3YjIyODQiLCJzdWIiOiJyZXBvOnRoZXB3YWduZXIvZ2l0aHViLXRva2VuLWFjdGlvbjpyZWY6cmVmcy9oZWFkcy9tYWluIiwiYXVkIjoiZ2l0aHViLXRva2VuLWFjdGlvbi1zZXJ2ZXIiLCJyZWYiOiJyZWZzL2hlYWRzL21haW4iLCJzaGEiOiIzYjkxM2I5MDRhZTZjZTc5YmU5OTU0NjlkMzAyYzRmODJjYWJiNzEyIiwicmVwb3NpdG9yeSI6InRoZXB3YWduZXIvZ2l0aHViLXRva2VuLWFjdGlvbiIsInJlcG9zaXRvcnlfb3duZXIiOiJ0aGVwd2FnbmVyIiwicmVwb3NpdG9yeV9vd25lcl9pZCI6IjE1NTk1MTAiLCJydW5faWQiOiIyNjc2NTEyNzUwIiwicnVuX251bWJlciI6IjUiLCJydW5fYXR0ZW1wdCI6IjEiLCJyZXBvc2l0b3J5X3Zpc2liaWxpdHkiOiJwcml2YXRlIiwicmVwb3NpdG9yeV9pZCI6IjUxNDIzODU5NiIsImFjdG9yX2lkIjoiMTU1OTUxMCIsImFjdG9yIjoidGhlcHdhZ25lciIsIndvcmtmbG93IjoiSGFjayIsImhlYWRfcmVmIjoiIiwiYmFzZV9yZWYiOiIiLCJldmVudF9uYW1lIjoicHVzaCIsInJlZl90eXBlIjoiYnJhbmNoIiwiam9iX3dvcmtmbG93X3JlZiI6InRoZXB3YWduZXIvZ2l0aHViLXRva2VuLWFjdGlvbi8uZ2l0aHViL3dvcmtmbG93cy9oYWNrLnltbEByZWZzL2hlYWRzL21haW4iLCJpc3MiOiJodHRwczovL3Rva2VuLmFjdGlvbnMuZ2l0aHVidXNlcmNvbnRlbnQuY29tIiwibmJmIjoxNjU3ODgzNTEyLCJleHAiOjE2NTc4ODQ0MTIsImlhdCI6MTY1Nzg4NDExMn0.1zr6A6Hc2sTZLMHRvJd925lG27ou0bAQoNkUnLY9RrRn-p7xSzqwlhm1wZxyOjILHSQDpgPy0COc2kQxCQPIusH2vxvuAOW7jrVphaKKC8aSamLe1Rs3VTec1u6bP_JDO2BWZuIcrnFd-yBlbyytXVlRw26Q6ZjkY7EoQXnHEmlskArcuFlTjjksaRRk2TPSnMyKBfWvrTNtVs1mN2CjzeX104NZ5qk-3m2jTInNTI-XJ7ZPlXsa0V12AzOzd9NdkdUe62Qk5LBN4W3XSUZWhS-nK3LjraRgKk_6JS8tFPKglsxUyYTpP3N8VPvYWRnjMzwRaJn1FagjwhSQ8K4NsQ`
	issuerGitHub = "https://token.actions.githubusercontent.com"
	issuerGoogle = "https://accounts.google.com"
)

func TestTokenParser_Actions(t *testing.T) {
	ctx := context.Background()
	ghTokenWasValid, err := time.Parse(time.RFC3339, "2022-07-15T11:26:30Z")
	require.NoError(t, err)
	tp, err := oidc.NewTokenParser(ctx, issuerGitHub, freezeTime(ghTokenWasValid))
	require.NoError(t, err)

	claims, err := tp.Parse(ctx, ghToken)
	require.NoError(t, err)

	assert.Equal(t, "push", claims["event_name"])
	assert.Equal(t, "thepwagner/github-token-action", claims["repository"])
	assert.Equal(t, "thepwagner", claims["actor"])
	assert.Equal(t, "refs/heads/main", claims["ref"])
	assert.Equal(t, "Hack", claims["workflow"])
}

// I'd stop the world and verify tokens with you
func freezeTime(t time.Time) oidc.TokenParserOpt {
	return func(cfg *coreoidc.Config) {
		cfg.Now = func() time.Time { return t }
	}
}
