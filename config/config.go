package config

import (
	"os"
)

type MCPConfig struct {
	Search SearchConfig
}

type SearchConfig struct {
	Provider string
	APIKey   string
	CX       string // For Google
}

func Load() (MCPConfig, error) {
	// Simple loader from env vars for now
	cfg := MCPConfig{
		Search: SearchConfig{
			Provider: os.Getenv("SEARCH_PROVIDER"),
			APIKey:   os.Getenv("SEARCH_API_KEY"),
			CX:       os.Getenv("SEARCH_CX"),
		},
	}
	return cfg, nil
}
