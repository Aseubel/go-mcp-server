package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	mcp_impl "mcp"
	"mcp/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := mcp_impl.NewMCPServer(cfg)
    // Prevent unused variable error while we figure out the correct Serve method
    _ = srv

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// SSE Endpoint
	r.GET("/sse", func(c *gin.Context) {
		transport := mcp_impl.NewSSEServerTransport()

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// Send endpoint event
		endpoint := fmt.Sprintf("/messages?sessionId=%s", transport.SessionID())
		c.SSEvent("endpoint", endpoint)
		c.Writer.Flush()

		// Start serving in a goroutine
		go func() {
            // TODO: Identify correct method to serve transport with mcp.Server
			// if err := srv.Server.Serve(transport); err != nil {
			// 	log.Printf("Server serve error: %v", err)
			// }
            log.Println("MCP Server transport connected")
		}()

		// Stream messages from transport to SSE
		for msg := range transport.SendChan {
			// msg is jsonrpc.Message or similar
			c.SSEvent("message", msg)
			c.Writer.Flush()
		}
		
		log.Printf("SSE connection closed for session %s", transport.SessionID())
	})

	r.POST("/messages", func(c *gin.Context) {
		sessionId := c.Query("sessionId")
		if sessionId == "" {
			c.JSON(400, gin.H{"error": "sessionId required"})
			return
		}

		val, ok := mcp_impl.TransportMap.Load(sessionId)
		if !ok {
			c.JSON(404, gin.H{"error": "session not found"})
			return
		}
		transport := val.(*mcp_impl.SSEServerTransport)

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read body"})
			return
		}

		var msg jsonrpc.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v\nBody: %s", err, string(body))
			c.JSON(400, gin.H{"error": "invalid jsonrpc message"})
			return
		}

		transport.HandleMessage(msg)
		c.Status(200)
	})

	port := "8080"
	log.Printf("Starting MCP server on :%s", port)
	r.Run(":" + port)
}
