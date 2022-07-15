package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v45/github"
	"github.com/thepwagner/github-token-action-server/api"
)

const (
	appID          = 89357
	installationID = 13062308
)

type TokenIssuer struct {
	log            logr.Logger
	gh             *github.Client
	installationID int64
}

func NewTokenIssuer(log logr.Logger) (*TokenIssuer, error) {
	log = log.WithName("TokenIssuer")
	tr, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, appID, "/Users/pwagner/Downloads/actions-secret-garden.2022-07-15.private-key.pem")
	if err != nil {
		return nil, fmt.Errorf("creating app transport: %w", err)
	}

	gh := github.NewClient(&http.Client{Transport: tr})
	return &TokenIssuer{
		log:            log,
		gh:             gh,
		installationID: installationID,
	}, nil
}

func (g *TokenIssuer) IssueToken(ctx context.Context, req *api.TokenRequest) (string, error) {
	tokReq := ConvertTokenRequest(req)
	g.log.Info("requesting token", "repositories", tokReq.Repositories, "permissions", tokReq.Permissions)
	tok, _, err := g.gh.Apps.CreateInstallationToken(ctx, g.installationID, tokReq)
	if err != nil {
		return "", fmt.Errorf("creating installation token: %w", err)
	}
	return tok.GetToken(), nil
}

func ConvertTokenRequest(req *api.TokenRequest) *github.InstallationTokenOptions {
	var opts github.InstallationTokenOptions
	for _, repo := range req.Repositories {
		repoSplit := strings.SplitN(repo, "/", 2)
		if len(repoSplit) == 2 {
			opts.Repositories = append(opts.Repositories, repoSplit[1])
		} else {
			opts.Repositories = append(opts.Repositories, repo)
		}
	}

	var perms github.InstallationPermissions
	for k, v := range req.Permissions {
		switch k {
		case "contents":
			perms.Contents = &v
		case "organization_projects":
			perms.OrganizationProjects = &v
			// TODO: the rest of this. come on copilot i don't want to
		}
	}

	opts.Permissions = &perms
	return &opts
}
