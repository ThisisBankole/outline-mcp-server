package config

import (
	"fmt"
	"os"
)

type Config struct {
	OutlineBaseURL  string
	OutlineAPIToken string
	MCPBearerToken  string
	Port            string
}

func Load() (*Config, error) {
	cfg := &Config{
		OutlineBaseURL:  os.Getenv("OUTLINE_BASE_URL"),
		OutlineAPIToken: os.Getenv("OUTLINE_API_TOKEN"),
		MCPBearerToken:  os.Getenv("MCP_BEARER_TOKEN"),
		Port:            os.Getenv("PORT"),
	}
	if cfg.OutlineBaseURL == "" {
		return nil, fmt.Errorf("OUTLINE_BASE_URL is required")
	}
	if cfg.OutlineAPIToken == "" {
		return nil, fmt.Errorf("OUTLINE_API_TOKEN is required")
	}
	if cfg.MCPBearerToken == "" {
		return nil, fmt.Errorf("MCP_BEARER_TOKEN is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg, nil
}
