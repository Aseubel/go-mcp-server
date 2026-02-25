package tools

import (
	"context"
	"fmt"
	"strings"

	"mcp/config"
	search_utils "mcp/tools/search"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchArgs 定义了搜索工具的输入参数结构
type SearchArgs struct {
	Query string `json:"query" jsonschema:"required"`
}

// SearchTool 实现了网络搜索工具
type SearchTool struct {
	Config   config.SearchConfig
	Provider search_utils.Provider
}

// NewSearchTool 创建一个新的 SearchTool
func NewSearchTool(cfg config.SearchConfig) *SearchTool {
	var provider search_utils.Provider
	switch strings.ToLower(cfg.Provider) {
	case "google":
		provider = &search_utils.GoogleProvider{APIKey: cfg.APIKey, CX: cfg.CX}
	case "serper":
		provider = &search_utils.SerperProvider{APIKey: cfg.APIKey}
	case "bocha":
		provider = &search_utils.BochaProvider{APIKey: cfg.APIKey}
	default:
		// 默认为 Tavily
		provider = &search_utils.TavilyProvider{APIKey: cfg.APIKey}
	}
	return &SearchTool{Config: cfg, Provider: provider}
}

// GetToolDef 返回 MCP 工具的具体定义
func (t *SearchTool) GetToolDef() *mcp.Tool {
	desc := fmt.Sprintf("使用 %s 在互联网上搜索实时信息。当你需要查找当前事件、事实或训练数据中不存在的最新信息时使用此工具。", t.Config.Provider)
	if t.Config.Provider == "" {
		desc = "使用 Google 搜索在互联网上搜索实时信息。当你需要查找当前事件、事实或训练数据中不存在的最新信息时使用此工具。"
	}

	return &mcp.Tool{
		Name:        "web_search",
		Description: desc,
		InputSchema: t.getInputSchema(),
	}
}

func (t *SearchTool) getInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "搜索查询字符串",
			},
		},
		"required": []string{"query"},
	}
}

// Register 将工具注册到 MCP Server
// 已弃用：请直接使用 GetToolDef 并在 server 中调用 RegisterTool
func (t *SearchTool) Register(s *mcp.Server) {
	mcp.AddTool(s, t.GetToolDef(), t.Execute)
}

// Execute 真正执行搜索逻辑
func (t *SearchTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	options := &search_utils.SearchOptions{
		MaxResults: 2,
	}

	items, err := t.Provider.Search(ctx, args.Query, options)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("执行搜索时出错: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	var sb strings.Builder
	if len(items) > 0 && items[0].Answer != "" {
		sb.WriteString("AI 回答: ")
		sb.WriteString(items[0].Answer)
		sb.WriteString("\n\n")
	}

	for i, item := range items {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item.Title))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", item.URL))
		sb.WriteString(fmt.Sprintf("   %s\n\n", item.Content))
	}

	if sb.Len() == 0 {
		sb.WriteString("未找到任何结果。")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
