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
		log.Fatal("无法加载配置", "error", err)
	}

	// 初始化日志系统
	log.Init(cfg)

	srv := mcp_impl.NewMCPServer(cfg)

	r := router.Setup(cfg, srv)

	port := fmt.Sprintf("%d", cfg.Server.Port)
	log.Info(fmt.Sprintf("正在启动 MCP 服务，基于端口 :%s", port))
	log.Info("  - Streamable HTTP 协议: POST /mcp (推荐)")
	log.Info("  - 传统 SSE 协议: GET /sse + POST /messages")
	r.Run(":" + port)
}
