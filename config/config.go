package config

import (
	"context"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Jira struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

type GitHub struct {
	Token string `yaml:"token"`
}

type Config struct {
	Jira   Jira   `yaml:"jira"`
	GitHub GitHub `yaml:"gitHub"`
}

func Load(ctx context.Context, filePath string) (*Config, error) {
	// Open config from file
	var cfg Config

	// Read the content of the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the yaml content into the Config struct
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
