package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/go-logr/logr"
)

// TokenParser validates a token and returns the identity of the Actions Workflow it was issued to.
type TokenParser func(ctx context.Context, tok string) (*WorkflowID, error)

type TokenIssuer interface {
	IssueToken(context.Context, *TokenRequest) (string, error)
}

type TokenResponse struct {
	Token string `json:"token"`
}

type Handler struct {
	log        logr.Logger
	parser     TokenParser
	authzCheck TokenCheck
	issuer     TokenIssuer
	requestID  uint64
}

func NewHandler(log logr.Logger, tp TokenParser, authzCheck TokenCheck, issuer TokenIssuer) *Handler {
	return &Handler{
		log:        log.WithName("Handler"),
		parser:     tp,
		authzCheck: authzCheck,
		issuer:     issuer,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := h.log.WithValues("request_id", atomic.AddUint64(&h.requestID, 1))
	log.V(1).Info("received request", "method", r.Method, "url", r.URL.String())

	workflow, err := h.authenticateWorkflow(r)
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.V(1).Info("parsed token request", "repositories", req.Repositories, "permissions", req.Permissions)

	if authorized, err := h.authzCheck(workflow, &req); err != nil {
		log.Error(err, "failed to authorize/deny token")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !authorized {
		http.Error(w, "token request denied", http.StatusForbidden)
		return
	}
	log.V(1).Info("authorized token")

	tok, err := h.issuer.IssueToken(r.Context(), &req)
	if err != nil {
		log.Error(err, "failed to issue token")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := TokenResponse{Token: tok}
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) authenticateWorkflow(r *http.Request) (*WorkflowID, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, fmt.Errorf("no authorization header")
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header")
	}
	tok := auth[len("Bearer "):]
	return h.parser(r.Context(), tok)
}
