package router

import (
	"github.com/gin-gonic/gin"

	mcp_impl "mcp"
	"mcp/internal/handler"
	"mcp/internal/middleware"
)

// Setup configures all routes for the MCP server.
func Setup(server *mcp_impl.MCPServer) *gin.Engine {
	r := gin.Default()

	// Global middleware
	r.Use(middleware.CORS())

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
