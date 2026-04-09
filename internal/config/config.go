package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Output struct {
		Default string `yaml:"default"`
	} `yaml:"output"`
	Task struct {
		DefaultProject string `yaml:"default_project"`
		InboxProjectID string `yaml:"inbox_project_id"`
	} `yaml:"task"`
	OAuth struct {
		ClientID    string `yaml:"client_id"`
		RedirectURL string `yaml:"redirect_url"`
	} `yaml:"oauth"`
}

type Store struct {
	Path string
}

func NewStore(path string) *Store {
	return &Store{Path: path}
}

func Default() Config {
	var cfg Config
	cfg.Output.Default = "table"
	return cfg
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tick", "config.yaml"), nil
}

func (s *Store) Load() (Config, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Config{}, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *Store) Save(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0o600)
}
