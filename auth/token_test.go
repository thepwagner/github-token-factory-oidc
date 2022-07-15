package auth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/require"
)

const tok = `eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6ImVCWl9jbjNzWFlBZDBjaDRUSEJLSElnT3dPRSIsImtpZCI6Ijc4MTY3RjcyN0RFQzVEODAxREQxQzg3ODRDNzA0QTFDODgwRUMwRTEifQ.eyJqdGkiOiI0YjMyZWExMi1iZDYwLTQ0ODEtOTZhOC0wZDU2ODE3YjIyODQiLCJzdWIiOiJyZXBvOnRoZXB3YWduZXIvZ2l0aHViLXRva2VuLWFjdGlvbjpyZWY6cmVmcy9oZWFkcy9tYWluIiwiYXVkIjoiZ2l0aHViLXRva2VuLWFjdGlvbi1zZXJ2ZXIiLCJyZWYiOiJyZWZzL2hlYWRzL21haW4iLCJzaGEiOiIzYjkxM2I5MDRhZTZjZTc5YmU5OTU0NjlkMzAyYzRmODJjYWJiNzEyIiwicmVwb3NpdG9yeSI6InRoZXB3YWduZXIvZ2l0aHViLXRva2VuLWFjdGlvbiIsInJlcG9zaXRvcnlfb3duZXIiOiJ0aGVwd2FnbmVyIiwicmVwb3NpdG9yeV9vd25lcl9pZCI6IjE1NTk1MTAiLCJydW5faWQiOiIyNjc2NTEyNzUwIiwicnVuX251bWJlciI6IjUiLCJydW5fYXR0ZW1wdCI6IjEiLCJyZXBvc2l0b3J5X3Zpc2liaWxpdHkiOiJwcml2YXRlIiwicmVwb3NpdG9yeV9pZCI6IjUxNDIzODU5NiIsImFjdG9yX2lkIjoiMTU1OTUxMCIsImFjdG9yIjoidGhlcHdhZ25lciIsIndvcmtmbG93IjoiSGFjayIsImhlYWRfcmVmIjoiIiwiYmFzZV9yZWYiOiIiLCJldmVudF9uYW1lIjoicHVzaCIsInJlZl90eXBlIjoiYnJhbmNoIiwiam9iX3dvcmtmbG93X3JlZiI6InRoZXB3YWduZXIvZ2l0aHViLXRva2VuLWFjdGlvbi8uZ2l0aHViL3dvcmtmbG93cy9oYWNrLnltbEByZWZzL2hlYWRzL21haW4iLCJpc3MiOiJodHRwczovL3Rva2VuLmFjdGlvbnMuZ2l0aHVidXNlcmNvbnRlbnQuY29tIiwibmJmIjoxNjU3ODgzNTEyLCJleHAiOjE2NTc4ODQ0MTIsImlhdCI6MTY1Nzg4NDExMn0.1zr6A6Hc2sTZLMHRvJd925lG27ou0bAQoNkUnLY9RrRn-p7xSzqwlhm1wZxyOjILHSQDpgPy0COc2kQxCQPIusH2vxvuAOW7jrVphaKKC8aSamLe1Rs3VTec1u6bP_JDO2BWZuIcrnFd-yBlbyytXVlRw26Q6ZjkY7EoQXnHEmlskArcuFlTjjksaRRk2TPSnMyKBfWvrTNtVs1mN2CjzeX104NZ5qk-3m2jTInNTI-XJ7ZPlXsa0V12AzOzd9NdkdUe62Qk5LBN4W3XSUZWhS-nK3LjraRgKk_6JS8tFPKglsxUyYTpP3N8VPvYWRnjMzwRaJn1FagjwhSQ8K4NsQ`

func Foo(ctx context.Context) error {
	prov, err := oidc.NewProvider(ctx, "https://token.actions.githubusercontent.com")
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	verifier := prov.Verifier(&oidc.Config{
		SkipClientIDCheck: true,
	})
	parsed, err := verifier.Verify(ctx, tok)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	fmt.Println(parsed)
	return nil
}

func TestToken(t *testing.T) {
	ctx := context.Background()
	err := Foo(ctx)
	require.NoError(t, err)

	fmt.Println(tok)
	t.Fail()
}
