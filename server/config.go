package server

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/thepwagner/github-token-factory-oidc/github"
)

type Config struct {
	Addr    string
	Issuers []string
	Checker CheckerConfig
	GitHub  map[string]github.Config
}

type CheckerConfig struct {
	Rego *RegoConfig
}

type RegoConfig struct {
	// If non-empty, an authoritative `.github/tokens.rego` will be loaded from this named repository.
	OwnerRepo string `mapstructure:"owner_repo"`
	// If set, `.github/tokens.rego` will be loaded from every repository in a request.
	FromRepos bool
}

// NewConfig loads config from the current directory.
func NewConfig() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.AddConfigPath(".")
	v.SetConfigName("gtfo")
	v.SetDefault("checker.rego.owner_repo", ".github")

	if err := v.ReadInConfig(); err != nil {
		var nfe viper.ConfigFileNotFoundError
		if !errors.As(err, &nfe) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}
