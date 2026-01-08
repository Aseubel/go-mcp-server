package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	mcp_impl "mcp"
)

// MCPHandler handles the Streamable HTTP MCP endpoint.
type MCPHandler struct {
	server *mcp_impl.MCPServer
}

// NewMCPHandler creates a new MCP handler.
func NewMCPHandler(server *mcp_impl.MCPServer) *MCPHandler {
	return &MCPHandler{server: server}
}

// jsonrpcRequest represents a JSON-RPC 2.0 request.
type jsonrpcRequest struct {
	JsonRpc string          `json:"jsonrpc"`
	Id      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// toolCallParams represents the parameters for tools/call method.
type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// Handle processes MCP requests via Streamable HTTP Transport.
// Single POST endpoint with SSE response.
func (h *MCPHandler) Handle(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to read body"})
		return
	}

	var request jsonrpcRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Printf("Failed to unmarshal request: %v\nBody: %s", err, string(body))
		c.JSON(400, gin.H{"error": "invalid jsonrpc message"})
		return
	}

	log.Printf("MCP Request: method=%s, id=%v", request.Method, request.Id)

	// Set SSE response headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	// Process MCP method
	response := h.processMethod(request)

	// Send SSE response
	responseBytes, _ := json.Marshal(response)
	fmt.Fprintf(c.Writer, "data: %s\n\n", responseBytes)
	c.Writer.Flush()
}

// processMethod routes the request to the appropriate handler.
func (h *MCPHandler) processMethod(request jsonrpcRequest) interface{} {
	switch request.Method {
	case "initialize":
		return h.handleInitialize(request.Id)
	case "tools/list":
		return h.handleToolsList(request.Id)
	case "tools/call":
		return h.handleToolsCall(request.Id, request.Params)
	default:
		return h.errorResponse(request.Id, -32601, "Method not found")
	}
}

// handleInitialize handles the initialize method.
func (h *MCPHandler) handleInitialize(id interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
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
}

// handleToolsList handles the tools/list method.
func (h *MCPHandler) handleToolsList(id interface{}) map[string]interface{} {
	tools := h.server.GetTools()
	toolList := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		toolList = append(toolList, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": toolList,
		},
	}
}

// handleToolsCall handles the tools/call method.
func (h *MCPHandler) handleToolsCall(id interface{}, params json.RawMessage) map[string]interface{} {
	var callParams toolCallParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return h.errorResponse(id, -32602, "Invalid params")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.server.CallTool(ctx, callParams.Name, callParams.Arguments)
	if err != nil {
		return h.errorResponse(id, -32603, err.Error())
	}

	// Convert MCP result to JSON
	content := make([]map[string]interface{}, 0)
	for _, c := range result.Content {
		content = append(content, map[string]interface{}{
			"type": "text",
			"text": fmt.Sprintf("%v", c),
		})
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"content": content,
			"isError": result.IsError,
		},
	}
}

// errorResponse creates a JSON-RPC error response.
func (h *MCPHandler) errorResponse(id interface{}, code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
}
