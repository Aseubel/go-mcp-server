package main

import (
	"fmt"
	"mcp/pkg/log"

	mcp_impl "mcp"
	"mcp/config"
	"mcp/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", "error", err)
	}

	// 初始化日志系统
	log.Init(cfg)

	srv := mcp_impl.NewMCPServer(cfg)

	r := router.Setup(cfg, srv)

	port := fmt.Sprintf("%d", cfg.Server.Port)
	log.Info(fmt.Sprintf("Starting MCP server on :%s", port))
	log.Info("  - Streamable HTTP: POST /mcp (recommended)")
	log.Info("  - Legacy SSE: GET /sse + POST /messages")
	r.Run(":" + port)
}
