package main

import (
	"log"

	mcp_impl "mcp"
	"mcp/config"
	"mcp/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := mcp_impl.NewMCPServer(cfg)

	r := router.Setup(srv)

	port := "8080"
	log.Printf("Starting MCP server on :%s", port)
	log.Printf("  - Streamable HTTP: POST /mcp (recommended)")
	log.Printf("  - Legacy SSE: GET /sse + POST /messages")
	r.Run(":" + port)
}
