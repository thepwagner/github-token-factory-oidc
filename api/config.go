package api

type Config struct {
	Addr   string
	GitHub map[string]GitHubConfig
}

type GitHubConfig struct {
	AppID          int64
	PrivateKeyPath string
}
