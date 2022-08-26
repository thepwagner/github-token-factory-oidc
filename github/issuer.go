package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v47/github"
	"github.com/thepwagner/github-token-factory-oidc/api"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Issuer struct {
	log    logr.Logger
	tracer trace.Tracer

	clients *Clients
}

func NewIssuer(log logr.Logger, tracer trace.Tracer, clients *Clients) *Issuer {
	return &Issuer{
		log:     log.WithName("github.Issuer"),
		tracer:  tracer,
		clients: clients,
	}
}

func (g *Issuer) IssueToken(ctx context.Context, req *api.TokenRequest) (string, error) {
	ctx, span := g.tracer.Start(ctx, "IssueToken")
	defer span.End()

	tokReq := ConvertTokenRequest(req)
	perms := make([]string, 0, len(req.Permissions))
	for k, v := range req.Permissions {
		perms = append(perms, fmt.Sprintf("%s:%s", k, v))
	}
	g.log.Info("requesting token", "repositories", tokReq.Repositories, "permissions", perms)
	span.SetAttributes(attribute.StringSlice("repositories", tokReq.Repositories), attribute.StringSlice("permissions", perms))

	client, err := g.clients.Client(ctx, req.Owner())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	tok, _, err := client.Apps.CreateInstallationToken(ctx, client.installationID, tokReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
		case "issues":
			perms.Issues = &v
		case "organization_projects":
			perms.OrganizationProjects = &v
			// TODO: the rest of this. come on copilot i don't want to
		}
	}

	opts.Permissions = &perms
	return &opts
}
