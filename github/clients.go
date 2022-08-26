package github

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v47/github"
)

type Config struct {
	AppID          int64  `mapstructure:"app_id"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
}

type Clients struct {
	transport  http.RoundTripper
	configs    map[string]Config
	clients    sync.Map
	appClients sync.Map
}

func NewClients(transport http.RoundTripper, configs map[string]Config) *Clients {
	return &Clients{
		transport: transport,
		configs:   configs,
	}
}

type Client struct {
	*github.Client
	installationID int64
}

func (c *Clients) Client(ctx context.Context, owner string) (*Client, error) {
	if client, ok := c.clients.Load(owner); ok {
		return client.(*Client), nil
	}

	// Load owner's config, fallback to default config
	cfg, ok := c.configs[owner]
	if !ok {
		cfg, ok = c.configs["*"]
		if !ok {
			return nil, fmt.Errorf("no configuration for repository owner %q", owner)
		}
	}
	tr, err := ghinstallation.NewAppsTransportKeyFromFile(c.transport, cfg.AppID, cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("creating app transport: %w", err)
	}
	client := github.NewClient(&http.Client{Transport: tr})

	// Assume the owner is an organization, fallback to user
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

	entry := &Client{
		Client:         client,
		installationID: installation.GetID(),
	}
	c.clients.Store(owner, entry)
	return entry, nil
}

func (c *Clients) AppClient(ctx context.Context, owner string) (*Client, error) {
	if client, ok := c.appClients.Load(owner); ok {
		return client.(*Client), nil
	}

	client, err := c.Client(ctx, owner)
	if err != nil {
		return nil, err
	}
	tr := client.Client.Client().Transport.(*ghinstallation.AppsTransport)
	transport := ghinstallation.NewFromAppsTransport(tr, client.installationID)

	entry := &Client{
		Client:         github.NewClient(&http.Client{Transport: transport}),
		installationID: client.installationID,
	}
	c.appClients.Store(owner, entry)
	return entry, nil
}
