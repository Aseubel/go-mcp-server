package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type MCPConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Search SearchConfig `mapstructure:"search"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type SearchConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	CX       string `mapstructure:"cx"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*MCPConfig, error) {
	v := viper.New()

	// Config file setup
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")        // Look in current directory
	v.AddConfigPath("./config") // Look in config directory

	// Default values
	v.SetDefault("server.port", 11611)
	v.SetDefault("server.env", "dev")
	v.SetDefault("search.provider", "bocha")

	// Log level
	v.SetDefault("log.level", "debug")

	// Environment variable setup
	// mapping: search.provider -> SEARCH_PROVIDER
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we fallback to env/defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg MCPConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}
