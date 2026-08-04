package router

import (
	"github.com/gin-gonic/gin"

	mcp_impl "mcp"
	"mcp/config"
	"mcp/internal/handler"
	"mcp/internal/middleware"
)

// NewRouter 构建基础 Gin 路由。公共/内部入口的认证策略在 Setup 中分别绑定。
func NewRouter(cfg *config.MCPConfig) *gin.Engine {
	// 设置Gin模式(Release/Debug)
	if cfg != nil && cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Recovery 是全局能力；身份认证必须按入口分别绑定。
	router.Use(gin.Recovery())

	return router
}

// Setup 为 MCP 服务器配置所有路由
func Setup(cfg *config.MCPConfig, server *mcp_impl.MCPServer) *gin.Engine {
	r := NewRouter(cfg)

	expectedServiceKey := ""
	if cfg != nil {
		expectedServiceKey = cfg.Server.AuthAPIKey
	}

	// 健康检查接口
	r.GET("/health", handler.Health)

	// Public MCP: only user-scoped memory tools are exposed externally.
	public := r.Group("")
	public.Use(middleware.RequireDeveloperKey())
	publicMCPHandler := handler.NewMCPHandler(server, mcp_impl.PublicToolSet)
	public.POST("/mcp", publicMCPHandler.Handle)

	// 传统 SSE 公共入口（兼容旧客户端）。
	publicSSEHandler := handler.NewSSEHandler(server, "/messages", "public")
	public.GET("/sse", publicSSEHandler.Connect)
	public.POST("/messages", publicSSEHandler.Message)

	// Internal MCP: Java uses the service key and sees only deployment-scoped
	// tools such as web_search. This route is not part of the public API.
	internal := r.Group("/internal")
	internal.Use(middleware.RequireServiceKey(expectedServiceKey))
	internalMCPHandler := handler.NewMCPHandler(server, mcp_impl.InternalToolSet)
	internal.POST("/mcp", internalMCPHandler.Handle)

	internalSSEHandler := handler.NewSSEHandler(server, "/internal/messages", "internal")
	internal.GET("/sse", internalSSEHandler.Connect)
	internal.POST("/messages", internalSSEHandler.Message)

	return r
}
