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
	// 设置Gin模式
	if cfg != nil && cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 中间件
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	return router
}

// Setup configures all routes for the MCP server.
func Setup(cfg *config.MCPConfig, server *mcp_impl.MCPServer) *gin.Engine {
	r := NewRouter(cfg)

	// Health check
	r.GET("/health", handler.Health)

	// Streamable HTTP Transport (recommended)
	mcpHandler := handler.NewMCPHandler(server)
	r.POST("/mcp", mcpHandler.Handle)

	// Legacy HTTP Transport (backward compatible)
	sseHandler := handler.NewSSEHandler(server)
	r.GET("/sse", sseHandler.Connect)
	r.POST("/messages", sseHandler.Message)

	return r
}
