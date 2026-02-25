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

// RegisterTool 将工具同时注册到 MCP server SDK 和内部注册表中
func RegisterTool[T any](s *MCPServer, tool *mcp.Tool, handler func(context.Context, *mcp.CallToolRequest, T) (*mcp.CallToolResult, any, error)) {
	// 向 SDK 注册
	mcp.AddTool(s.Server, tool, handler)

	// 内部注册
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
			// 我们对 CallToolRequest 传 nil，因为这是一个内部调用
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

	// 注册工具
	if cfg.Search.Provider != "" {
		searchTool := tools.NewSearchTool(cfg.Search)
		RegisterTool(mcpSrv, searchTool.GetToolDef(), searchTool.Execute)
	}

	// 初始化后端 gRPC 连接
	grpcTarget := cfg.Grpc.BackendTarget
	grpc.InitClient(grpcTarget)

	// 注册扩展工具 (Extension Tools)
	diarySearchTool := ext_tools.NewSearchDiaryTool()
	RegisterTool(mcpSrv, diarySearchTool.GetToolDef(), diarySearchTool.Execute)

	lifeGraphTool := ext_tools.NewQueryLifeGraphTool()
	RegisterTool(mcpSrv, lifeGraphTool.GetToolDef(), lifeGraphTool.Execute)

	return mcpSrv
}

// CallTool 根据工具名称执行已注册的工具
func (s *MCPServer) CallTool(ctx context.Context, name string, argsJSON []byte) (*mcp.CallToolResult, error) {
	tool, ok := s.Tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool.Handler(ctx, argsJSON)
}

// GetTools 返回所有已注册工具的定义
func (s *MCPServer) GetTools() []*mcp.Tool {
	tools := make([]*mcp.Tool, 0, len(s.Tools))
	for _, t := range s.Tools {
		tools = append(tools, t.Tool)
	}
	return tools
}
