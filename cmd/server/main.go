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
	log.Info("  - 公共 MCP: POST /mcp (diarySearch, memorySearch)")
	log.Info("  - 内部 MCP: POST /internal/mcp (web_search, service key)")
	log.Info("  - 公共 SSE: GET /sse + POST /messages")
	log.Info("  - 内部 SSE: GET /internal/sse + POST /internal/messages")
	r.Run(":" + port)
}
