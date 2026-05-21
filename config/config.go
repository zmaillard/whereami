package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DbPath        string `envconfig:"DB_PATH"`
	VersionNumber string
}

func NewConfig() (*Config, error) {
	return NewConfigWithVersion("development")
}

func NewConfigWithVersion(version string) (*Config, error) {
	var config Config
	err := envconfig.Process("", &config)
	config.VersionNumber = version
	return &config, err
}
