package checker

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/thepwagner/github-token-factory-oidc/api"
	"github.com/thepwagner/github-token-factory-oidc/github"
	"golang.org/x/sync/errgroup"
)

type RepoRego struct {
	log    logr.Logger
	github *github.Clients
}

func NewRepoRego(log logr.Logger, github *github.Clients) *RepoRego {
	return &RepoRego{
		log:    log.WithName("auth.RepoRego"),
		github: github,
	}
}

var _ api.TokenChecker = (*RepoRego)(nil)

func (r RepoRego) Check(ctx context.Context, claims api.Claims, req *api.TokenRequest) (bool, error) {
	// Fetch all the relevant policies:
	ownerCheck, repoChecks, err := r.fetchRepoPolicies(ctx, req)
	if err != nil {
		return false, err
	}

	// If a policy is set by the owner, it can approve:
	if ownerCheck != nil {
		ownerOk, err := ownerCheck.Check(ctx, claims, req)
		if err != nil {
			return false, fmt.Errorf("checking owner policy: %w", err)
		}
		r.log.Info("evaluated owner policy", "ok", ownerOk)
		if ownerOk {
			return true, nil
		}
	}
	// FIXME: org-level permissions can only be granted by the owner
	if req.OwnerPermissions() {
		return false, nil
	}

	// Otherwise, every repository policy must approve:
	if len(repoChecks) != len(req.Repositories) {
		return false, fmt.Errorf("expected %d repo policies, got %d", len(req.Repositories), len(repoChecks))
	}
	for _, repoCheck := range repoChecks {
		repoOk, err := repoCheck.Check(ctx, claims, req)
		if err != nil {
			return false, fmt.Errorf("checking repository policy: %w", err)
		}
		r.log.Info("evaluated repo policy", "ok", repoOk)
		if !repoOk {
			return false, nil
		}
	}

	return true, nil
}

func (r RepoRego) fetchRepoPolicies(ctx context.Context, req *api.TokenRequest) (*Rego, []*Rego, error) {
	eg, ctx := errgroup.WithContext(ctx)
	requested := make(map[string]struct{}, len(req.Repositories))
	repoRegos := make(chan *Rego, 1)
	for _, repo := range req.Repositories {
		repo := repo
		if _, ok := requested[repo]; ok {
			continue
		}
		requested[repo] = struct{}{}

		eg.Go(r.fetchRepoPolicy(ctx, repo, repoRegos))
	}
	ownerRego := make(chan *Rego, 1)
	ownerRepo := fmt.Sprintf("%s/.github", req.Owner())
	if _, ok := requested[ownerRepo]; !ok {
		eg.Go(r.fetchRepoPolicy(ctx, ownerRepo, ownerRego))
	}
	go func() {
		_ = eg.Wait()
		close(repoRegos)
		close(ownerRego)
	}()

	// Collect results:
	repoChecks := make([]*Rego, 0, len(requested))
	for c := range repoRegos {
		repoChecks = append(repoChecks, c)
	}
	ownerCheck := <-ownerRego
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}
	return ownerCheck, repoChecks, nil
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
