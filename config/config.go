package config

import "os"

type MCPConfig struct {
	Search SearchConfig
}

type SearchConfig struct {
	Provider string
	APIKey   string
	CX       string // For Google
}

func getEnvDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func Load() (MCPConfig, error) {
	// 从环境变量读取，若不存在则使用默认值
	cfg := MCPConfig{
		Search: SearchConfig{
			Provider: getEnvDefault("SEARCH_PROVIDER", "bocha"),
			APIKey:   getEnvDefault("SEARCH_API_KEY", ""),
			CX:       getEnvDefault("SEARCH_CX", ""),
		},
	}
	return cfg, nil
}
