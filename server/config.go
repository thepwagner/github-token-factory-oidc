package server

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/thepwagner/github-token-action-server/github"
)

type Config struct {
	Addr           string
	JaegerEndpoint string `mapstructure:"jaeger_endpoint"`
	Issuers        []string
	GitHub         map[string]github.IssuerConfig
}

func NewConfig() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	v.AddConfigPath(".")
	v.SetConfigName("gtfo")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}
