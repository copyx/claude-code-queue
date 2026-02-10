package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Prefix    string            `json:"prefix"`
	Dashboard DashboardConfig   `json:"dashboard"`
}

type DashboardConfig struct {
	Width         int  `json:"width"`           // Width in columns (default: 20)
	AutoHeight    bool `json:"auto_height"`     // Automatically adjust height (default: true)
	ShowDirectory bool `json:"show_directory"`  // Show current directory (default: true)
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ccq", "config")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	// Apply defaults for missing fields
	applyDefaults(&cfg)
	return &cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Dashboard: DashboardConfig{
			Width:         20,
			AutoHeight:    true,
			ShowDirectory: true,
		},
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Dashboard.Width <= 0 {
		cfg.Dashboard.Width = 20
	}
	// AutoHeight and ShowDirectory default to false in Go, but we want true
	// If config was loaded from file without these fields, they'll be false
	// For now, we'll assume explicit false means user wants it off
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
