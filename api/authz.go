package api

import (
	"context"
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

func (r TokenRequest) Owner() string {
	if len(r.Repositories) == 0 {
		return ""
	}
	return strings.Split(r.Repositories[0], "/")[0]
}

func (r TokenRequest) OwnerPermissions() bool {
	for perm := range r.Permissions {
		switch perm {
		case "organization_projects":
			return true
		}
	}
	return false
}

// TokenCheck checks if a client is authorized to request a token
type TokenChecker interface {
	Check(context.Context, Claims, *TokenRequest) (bool, error)
}
