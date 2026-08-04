package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type MCPConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Search SearchConfig `mapstructure:"search"`
	Grpc   GrpcConfig   `mapstructure:"grpc"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port       int    `mapstructure:"port"`
	Env        string `mapstructure:"env"`
	AuthAPIKey string `mapstructure:"auth_api_key"`
}

type SearchConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	CX       string `mapstructure:"cx"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type GrpcConfig struct {
	BackendTarget string `mapstructure:"backend_target"`
}

func Load() (*MCPConfig, error) {
	v := viper.New()

	// 配置文件的寻找路径和名称
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")        // 在当前目录寻找
	v.AddConfigPath("./config") // 在 config 目录寻找

	// 设置默认值
	v.SetDefault("server.port", 11611)
	v.SetDefault("server.env", "dev")
	v.SetDefault("server.auth_api_key", "")
	v.SetDefault("search.provider", "bocha")
	v.SetDefault("grpc.backend_target", "localhost:9090")

	// 日志级别默认值
	v.SetDefault("log.level", "debug")

	// 环境变量支持
	// 将诸如 search.provider 映射为 SEARCH_PROVIDER 环境变量
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	for key, envName := range map[string]string{
		"server.port":         "SERVER_PORT",
		"server.env":          "SERVER_ENV",
		"server.auth_api_key": "SERVER_AUTH_API_KEY",
		"search.provider":     "SEARCH_PROVIDER",
		"search.api_key":      "SEARCH_API_KEY",
		"search.cx":           "SEARCH_CX",
		"grpc.backend_target": "GRPC_BACKEND_TARGET",
		"log.level":           "LOG_LEVEL",
	} {
		if err := v.BindEnv(key, envName); err != nil {
			return nil, fmt.Errorf("绑定环境变量 %s 失败: %w", envName, err)
		}
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在是可以接受的配置，我们降级使用环境变量和默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件时发生错误: %w", err)
		}
	}

	var cfg MCPConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("无法将解析出的配置映射为结构体: %w", err)
	}
	if strings.TrimSpace(cfg.Server.AuthAPIKey) == "" {
		return nil, fmt.Errorf("server.auth_api_key is required")
	}

	return &cfg, nil
}
