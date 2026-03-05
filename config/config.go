package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	MaxTokens   int32   `yaml:"max_tokens"`
	Temperature float32 `yaml:"temperature"`
}

func DefaultConfig() *Config {
	return &Config{
		Model:       "gemini-2.0-flash",
		MaxTokens:   8192,
		Temperature: 0.2,
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Try loading from ~/.openclaw.yaml
	home, err := os.UserHomeDir()
	if err == nil {
		data, err := os.ReadFile(home + "/.openclaw.yaml")
		if err == nil {
			_ = yaml.Unmarshal(data, cfg)
		}
	}

	// Env vars override file config
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set. Set it via environment variable or ~/.openclaw.yaml")
	}

	return cfg, nil
}
