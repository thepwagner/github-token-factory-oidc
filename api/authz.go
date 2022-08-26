package api

import (
	"fmt"
	"strings"
)

type Claims map[string]interface{}

// TokenRequest is a request from a workflow for permissions
type TokenRequest struct {
	Repositories []string          `json:"repositories"`
	Permissions  map[string]string `json:"permissions"`
}

func (r TokenRequest) Valid() error {
	if len(r.Repositories) == 0 {
		return fmt.Errorf("no repositories")
	}
	if len(r.Permissions) == 0 {
		return fmt.Errorf("no permissions")
	}
	return nil
}

func (r TokenRequest) Organization() string {
	if len(r.Repositories) == 0 {
		return ""
	}
	return strings.Split(r.Repositories[0], "/")[0]
}

// TokenCheck checks if a workflow is authorized to request a token
type TokenCheck func(Claims, *TokenRequest) (bool, error)

func TokenCheckYOLO(c Claims, _ *TokenRequest) (bool, error) {
	fmt.Printf("%+v\n", c)
	return true, nil
}
