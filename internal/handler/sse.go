package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	mcp_impl "mcp"
)

// SSEHandler handles the legacy SSE transport endpoints.
type SSEHandler struct {
	server *mcp_impl.MCPServer
}

// NewSSEHandler creates a new SSE handler.
func NewSSEHandler(server *mcp_impl.MCPServer) *SSEHandler {
	return &SSEHandler{server: server}
}

// Connect handles GET /sse - establishes SSE connection.
func (h *SSEHandler) Connect(c *gin.Context) {
	transport := mcp_impl.NewSSEServerTransport()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Send endpoint event
	endpoint := fmt.Sprintf("/messages?sessionId=%s", transport.SessionID())
	c.SSEvent("endpoint", endpoint)
	c.Writer.Flush()

	// Start serving in a goroutine
	go func() {
		apiKey := c.GetString("apiKey")
		log.Printf("MCP Server transport connected (legacy SSE), apiKey=%s", apiKey)
		// 这里的 Server 应该启动来处理 transport 上的消息
		// h.server.Server.ServeStream(transport) 
	}()

	// Stream messages from transport to SSE
	for msg := range transport.SendChan {
		c.SSEvent("message", msg)
		c.Writer.Flush()
	}

	log.Printf("SSE connection closed for session %s", transport.SessionID())
}

// Message handles POST /messages - receives messages for a session.
func (h *SSEHandler) Message(c *gin.Context) {
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
}
