package config

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Output         string `yaml:"output"`
	Region         string `yaml:"region"`
	DefaultProject string `yaml:"default_project"`
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
	RedirectURL    string `yaml:"redirect_url"`
}

type Store struct {
	Path string
}

func NewStore(path string) *Store {
	return &Store{Path: path}
}

func Default() Config {
	var cfg Config
	cfg.Output = "table"
	cfg.Region = "ticktick"
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
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		if errors.Is(err, io.EOF) {
			return cfg, nil
		}
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
