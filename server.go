package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"mcp/config"
	"mcp/internal/grpc"
	ext_tools "mcp/internal/tools"
	"mcp/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type InternalToolHandler func(ctx context.Context, argsJSON json.RawMessage) (*mcp.CallToolResult, error)

type RegisteredTool struct {
	Tool    *mcp.Tool
	Handler InternalToolHandler
}

type MCPServer struct {
	Server *mcp.Server
	Config *config.MCPConfig
	Tools  map[string]RegisteredTool
}

// RegisterTool registers a tool with both the MCP server SDK and the internal registry
func RegisterTool[T any](s *MCPServer, tool *mcp.Tool, handler func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error)) {
	// Register with SDK
	mcp.AddTool(s.Server, tool, handler)

	// Register internally
	if s.Tools == nil {
		s.Tools = make(map[string]RegisteredTool)
	}
	s.Tools[tool.Name] = RegisteredTool{
		Tool: tool,
		Handler: func(ctx context.Context, argsJSON json.RawMessage) (*mcp.CallToolResult, error) {
			var args T
			if err := json.Unmarshal(argsJSON, &args); err != nil {
				return nil, fmt.Errorf("failed to unmarshal args: %w", err)
			}
			// We pass nil for CallToolRequest as it's an internal call
			res, _, err := handler(ctx, nil, args)
			return res, err
		},
	}
}

func NewMCPServer(cfg *config.MCPConfig) *MCPServer {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "ai-ability-mcp",
		Version: "1.0.0",
	}, nil)

	mcpSrv := &MCPServer{
		Server: s,
		Config: cfg,
		Tools:  make(map[string]RegisteredTool),
	}

	// Register Tools
	if cfg.Search.Provider != "" {
		searchTool := tools.NewSearchTool(cfg.Search)
		RegisterTool(mcpSrv, searchTool.GetToolDef(), searchTool.Execute)
	}

	// Initialize backend gRPC connection
	grpcTarget := cfg.Grpc.BackendTarget
	grpc.InitClient(grpcTarget)

	// Register Extension Tools
	diarySearchTool := ext_tools.NewSearchDiaryTool()
	RegisterTool(mcpSrv, diarySearchTool.GetToolDef(), diarySearchTool.Execute)

	lifeGraphTool := ext_tools.NewQueryLifeGraphTool()
	RegisterTool(mcpSrv, lifeGraphTool.GetToolDef(), lifeGraphTool.Execute)

	return mcpSrv
}

// CallTool executes a registered tool by name
func (s *MCPServer) CallTool(ctx context.Context, name string, argsJSON []byte) (*mcp.CallToolResult, error) {
	tool, ok := s.Tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool.Handler(ctx, argsJSON)
}

// GetTools returns all registered tools definitions
func (s *MCPServer) GetTools() []*mcp.Tool {
	tools := make([]*mcp.Tool, 0, len(s.Tools))
	for _, t := range s.Tools {
		tools = append(tools, t.Tool)
	}
	return tools
}
