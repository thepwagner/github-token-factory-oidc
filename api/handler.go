package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TokenParser validates a token and returns the identity of the Actions Workflow it was issued to.
type TokenParser interface {
	Parse(ctx context.Context, tok string) (Claims, error)
}

type TokenIssuer func(context.Context, *TokenRequest) (string, error)

type TokenResponse struct {
	Token     string `json:"token"`
	Revocable bool   `json:"revocable"`
	Error     string `json:"error,omitempty"`
}

type Handler struct {
	log    *slog.Logger
	tracer trace.Tracer

	tokenParser  TokenParser
	tokenChecker TokenChecker
	tokenIssuer  TokenIssuer
}

func NewHandler(log *slog.Logger, tracer trace.Tracer, tokenParser TokenParser, tokenChecker TokenChecker, tokenIssuer TokenIssuer) *Handler {
	return &Handler{
		log:          log.With("logger", "Handler"),
		tracer:       tracer,
		tokenParser:  tokenParser,
		tokenChecker: tokenChecker,
		tokenIssuer:  tokenIssuer,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	ctx, span := h.tracer.Start(r.Context(), "handler.ServeHTTP")
	defer span.End()

	tok, status, err := h.tokenRequest(ctx, r)
	resp := TokenResponse{
		Token:     tok,
		Revocable: true,
	}
	if err != nil {
		h.log.Error("error issuing token", slog.String("err", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) tokenRequest(ctx context.Context, r *http.Request) (string, int, error) {
	h.log.Debug("received request", "url", r.URL.String())

	claims, err := h.authenticate(ctx, r)
	if err != nil {
		return "", http.StatusUnauthorized, err
	}

	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", http.StatusBadRequest, err
	} else if err := req.Valid(); err != nil {
		return "", http.StatusBadRequest, err
	}

	if authorized, err := h.tokenChecker.Check(ctx, claims, &req); err != nil {
		return "", http.StatusInternalServerError, err
	} else if !authorized {
		return "", http.StatusForbidden, fmt.Errorf("not authorized")
	}
	h.log.Debug("authorized token")

	tok, err := h.tokenIssuer(ctx, &req)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	return tok, http.StatusOK, nil
}

func (h *Handler) authenticate(ctx context.Context, r *http.Request) (Claims, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, fmt.Errorf("no authorization header")
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("invalid authorization header")
	}
	tok := auth[len("Bearer "):]
	return h.tokenParser.Parse(ctx, tok)
}
