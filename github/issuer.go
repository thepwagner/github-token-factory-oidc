package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-logr/logr"
	"github.com/google/go-github/v45/github"
	"github.com/thepwagner/github-token-action-server/api"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type IssuerConfig struct {
	AppID          int64  `mapstructure:"app_id"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
}

type Issuer struct {
	log    logr.Logger
	tracer trace.Tracer

	transport http.RoundTripper
	configs   map[string]IssuerConfig
	clients   sync.Map
}

type clientEntry struct {
	client *github.Client
	id     int64
}

func NewIssuer(log logr.Logger, tp trace.TracerProvider, configs map[string]IssuerConfig) *Issuer {
	return &Issuer{
		log:       log.WithName("github.Issuer"),
		tracer:    tp.Tracer(""),
		transport: otelhttp.NewTransport(http.DefaultTransport, otelhttp.WithTracerProvider(tp)),
		configs:   configs,
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

	installation, err := g.newGitHubClient(ctx, req.Organization())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	tok, _, err := installation.client.Apps.CreateInstallationToken(ctx, installation.id, tokReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", fmt.Errorf("creating installation token: %w", err)
	}
	return tok.GetToken(), nil
}

func (g *Issuer) newGitHubClient(ctx context.Context, owner string) (*clientEntry, error) {
	if client, ok := g.clients.Load(owner); ok {
		return client.(*clientEntry), nil
	}

	cfg, ok := g.configs[owner]
	if !ok {
		cfg, ok = g.configs["*"]
		if !ok {
			return nil, fmt.Errorf("no configuration for repository owner %q", owner)
		}
	}

	tr, err := ghinstallation.NewAppsTransportKeyFromFile(g.transport, cfg.AppID, cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("creating app transport: %w", err)
	}
	client := github.NewClient(&http.Client{Transport: tr})

	installation, res, err := client.Apps.FindOrganizationInstallation(ctx, owner)
	if err != nil {
		if res.StatusCode != http.StatusNotFound {
			return nil, fmt.Errorf("finding org installation: %w", err)
		}

		installation, _, err = client.Apps.FindUserInstallation(ctx, owner)
		if err != nil {
			return nil, fmt.Errorf("finding org+user installation: %w", err)
		}
	}

	entry := &clientEntry{
		client: client,
		id:     installation.GetID(),
	}
	g.clients.Store(owner, entry)
	return entry, nil
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
