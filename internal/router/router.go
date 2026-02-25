package router

import (
	"github.com/gin-gonic/gin"

	mcp_impl "mcp"
	"mcp/config"
	"mcp/internal/handler"
	"mcp/internal/middleware"
)

// NewRouter 构建带中间件与全部路由的Gin路由
func NewRouter(cfg *config.MCPConfig) *gin.Engine {
	// 设置Gin模式(Release/Debug)
	if cfg != nil && cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 添加全局中间件: Recovery 与 CORS
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	return router
}

// Setup 为 MCP 服务器配置所有路由
func Setup(cfg *config.MCPConfig, server *mcp_impl.MCPServer) *gin.Engine {
	r := NewRouter(cfg)

	// 健康检查接口
	r.GET("/health", handler.Health)

	// Streamable HTTP 通讯协议路由 (官方推荐)
	mcpHandler := handler.NewMCPHandler(server)
	r.POST("/mcp", mcpHandler.Handle)

	// 传统 SSE 通讯协议路由 (为了向下兼容)
	sseHandler := handler.NewSSEHandler(server)
	r.GET("/sse", sseHandler.Connect)
	r.POST("/messages", sseHandler.Message)

	return r
}
