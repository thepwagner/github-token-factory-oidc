package checker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/thepwagner/github-token-factory-oidc/api"
	"github.com/thepwagner/github-token-factory-oidc/github"
	"golang.org/x/sync/errgroup"
)

type RepoRego struct {
	log       *slog.Logger
	github    *github.Clients
	ownerRepo string
	everyRepo bool
}

func NewRepoRego(log *slog.Logger, github *github.Clients, ownerRepo string, everyRepo bool) *RepoRego {
	return &RepoRego{
		log:       log.With("logger", "auth.RepoRego"),
		github:    github,
		ownerRepo: ownerRepo,
		everyRepo: everyRepo,
	}
}

var _ api.TokenChecker = (*RepoRego)(nil)

func (r RepoRego) Check(ctx context.Context, claims api.Claims, req *api.TokenRequest) (bool, error) {
	// The token may be approved by the global "owner" policy
	if ownerOk, err := r.checkOwnerPolicy(ctx, claims, req); err != nil {
		return false, err
	} else if ownerOk {
		return true, nil
	}

	// For permissions that affect the owner (not individiual repos), don't listen to repositories
	if req.OwnerPermissions() {
		return false, nil
	}

	return r.checkRepoPolicies(ctx, claims, req)
}

func (r RepoRego) checkOwnerPolicy(ctx context.Context, claims api.Claims, req *api.TokenRequest) (bool, error) {
	if r.ownerRepo == "" {
		// No owner policy is configured
		return false, nil
	}

	ownerRepo := r.resolveOwnerRepo(req)
	rego, err := r.fetchRepoPolicies(ctx, ownerRepo)
	if err != nil {
		return false, fmt.Errorf("fetching owner policy: %w", err)
	}
	if len(rego) == 0 {
		// No policy found in the configured repository
		return false, nil
	}

	res, err := rego[0].Check(ctx, claims, req)
	if err != nil {
		return false, fmt.Errorf("checking owner policy: %w", err)
	}
	r.log.Info("evaluated owner policy", "ok", res, "owner_repo", ownerRepo)
	return res, nil
}

func (r RepoRego) checkRepoPolicies(ctx context.Context, claims api.Claims, req *api.TokenRequest) (bool, error) {
	if !r.everyRepo {
		return false, nil
	}

	ownerRepo := r.resolveOwnerRepo(req)
	toFetch := make([]string, 0, len(req.Repositories))
	uniq := make(map[string]struct{}, len(req.Repositories))
	for _, repo := range req.Repositories {
		// We know the owner repo will fail, ABORT!
		if repo == ownerRepo {
			return false, nil
		}
		if _, ok := uniq[repo]; ok {
			continue
		}
		uniq[repo] = struct{}{}
		toFetch = append(toFetch, repo)
	}

	regos, err := r.fetchRepoPolicies(ctx, toFetch...)
	if err != nil {
		return false, fmt.Errorf("fetching repository policies: %w", err)
	}
	if len(regos) != len(toFetch) {
		return false, fmt.Errorf("expected %d repo policies, got %d", len(toFetch), len(regos))
	}
	for _, rego := range regos {
		res, err := rego.Check(ctx, claims, req)
		if err != nil {
			return false, fmt.Errorf("checking repository policy: %w", err)
		}
		r.log.Info("evaluated repo policy", "ok", res)
		if !res {
			// Must by accepted by every policy, so the first rejection is terminal:
			return false, nil
		}
	}

	// The request has been approved by all repository policies
	return true, nil
}

func (r RepoRego) resolveOwnerRepo(req *api.TokenRequest) string {
	if strings.Contains(r.ownerRepo, "/") {
		// Repository is owner/name
		return r.ownerRepo
	}

	// Repository is just a name - combine with the owner of repositories in the request
	return fmt.Sprintf("%s/%s", req.Owner(), r.ownerRepo)
}

func (r RepoRego) fetchRepoPolicies(ctx context.Context, repos ...string) ([]*Rego, error) {
	eg, ctx := errgroup.WithContext(ctx)
	regos := make(chan *Rego, 1)
	for _, repo := range repos {
		repo := repo
		eg.Go(r.fetchRepoPolicy(ctx, repo, regos))
	}
	go func() {
		_ = eg.Wait()
		close(regos)
	}()

	// Collect results:
	res := make([]*Rego, 0, len(repos))
	for c := range regos {
		res = append(res, c)
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func (r RepoRego) fetchRepoPolicy(ctx context.Context, repo string, res chan *Rego) func() error {
	return func() error {
		repoParts := strings.Split(repo, "/")
		if len(repoParts) != 2 {
			return fmt.Errorf("invalid repo: %s", repo)
		}

		client, err := r.github.AppClient(ctx, repoParts[0])
		if err != nil {
			return fmt.Errorf("getting client for %s: %w", repo, err)
		}
		fc, _, _, err := client.Repositories.GetContents(ctx, repoParts[0], repoParts[1], ".github/tokens.rego", nil)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "404 not found") {
				return nil
			}
			return fmt.Errorf("fetching repo policy: %w", err)
		}

		r.log.Info("fetched policy", "repository", repo, "sha", fc.GetSHA())
		policyRaw, err := fc.GetContent()
		if err != nil {
			return fmt.Errorf("fetching repo policy content: %w", err)
		}

		rego, err := NewRego(ctx, r.log, policyRaw)
		if err != nil {
			return fmt.Errorf("parsing repo policy: %w", err)
		}
		res <- rego
		return nil
	}
}
