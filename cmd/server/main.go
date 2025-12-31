package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

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

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ============================================================
	// Streamable HTTP Transport (新版推荐)
	// 单一 POST 端点，响应为 SSE 流
	// ============================================================
	r.POST("/mcp", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read body"})
			return
		}

		// 解析 JSON-RPC 请求
		var request struct {
			JsonRpc string          `json:"jsonrpc"`
			Id      interface{}     `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params,omitempty"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			log.Printf("Failed to unmarshal request: %v\nBody: %s", err, string(body))
			c.JSON(400, gin.H{"error": "invalid jsonrpc message"})
			return
		}

		log.Printf("MCP Request: method=%s, id=%v", request.Method, request.Id)

		// 设置 SSE 响应头
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Flush()

		// 处理 MCP 方法
		var response interface{}
		switch request.Method {
		case "initialize":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      request.Id,
				"result": map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities": map[string]interface{}{
						"tools": map[string]interface{}{},
					},
					"serverInfo": map[string]interface{}{
						"name":    "yusi-mcp-server",
						"version": "1.0.0",
					},
				},
			}
		case "tools/list":
			// 获取所有注册的工具
			tools := srv.GetTools()
			toolList := make([]map[string]interface{}, 0, len(tools))
			for _, tool := range tools {
				toolList = append(toolList, map[string]interface{}{
					"name":        tool.Name,
					"description": tool.Description,
					"inputSchema": tool.InputSchema,
				})
			}
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      request.Id,
				"result": map[string]interface{}{
					"tools": toolList,
				},
			}
		case "tools/call":
			// 解析工具调用参数
			var callParams struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if err := json.Unmarshal(request.Params, &callParams); err != nil {
				response = map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      request.Id,
					"error": map[string]interface{}{
						"code":    -32602,
						"message": "Invalid params",
					},
				}
			} else {
				// 执行工具
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				
				result, err := srv.CallTool(ctx, callParams.Name, callParams.Arguments)
				if err != nil {
					response = map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      request.Id,
						"error": map[string]interface{}{
							"code":    -32603,
							"message": err.Error(),
						},
					}
				} else {
					// 转换 MCP result 到 JSON
					content := make([]map[string]interface{}, 0)
					for _, c := range result.Content {
						content = append(content, map[string]interface{}{
							"type": "text",
							"text": fmt.Sprintf("%v", c),
						})
					}
					response = map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      request.Id,
						"result": map[string]interface{}{
							"content": content,
							"isError": result.IsError,
						},
					}
				}
			}
		default:
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      request.Id,
				"error": map[string]interface{}{
					"code":    -32601,
					"message": "Method not found",
				},
			}
		}

		// 发送 SSE 响应
		responseBytes, _ := json.Marshal(response)
		fmt.Fprintf(c.Writer, "data: %s\n\n", responseBytes)
		c.Writer.Flush()
	})

	// ============================================================
	// Legacy HTTP Transport (向后兼容)
	// GET /sse + POST /messages
	// ============================================================
	r.GET("/sse", func(c *gin.Context) {
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
			log.Println("MCP Server transport connected (legacy SSE)")
		}()

		// Stream messages from transport to SSE
		for msg := range transport.SendChan {
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
	log.Printf("  - Streamable HTTP: POST /mcp (recommended)")
	log.Printf("  - Legacy SSE: GET /sse + POST /messages")
	r.Run(":" + port)
}

