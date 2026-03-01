package tools

import (
	"context"
	"fmt"

	"mcp/internal/grpc"
	pb "mcp/proto"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchDiaryArgs 定义了 diarySearch 工具的入参结构
type SearchDiaryArgs struct {
	Keyword   string `json:"keyword"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

// SearchDiaryTool 用于执行后端的 SearchDiary gRPC 方法
type SearchDiaryTool struct{}

// NewSearchDiaryTool 创建一个新的 SearchDiaryTool 实例
func NewSearchDiaryTool() *SearchDiaryTool {
	return &SearchDiaryTool{}
}

// GetToolDef 返回该工具在 MCP 中注册的定义
func (t *SearchDiaryTool) GetToolDef() *mcp.Tool {
	return &mcp.Tool{
		Name:        "diarySearch",
		Description: "根据关键词和可选的时间范围搜索用户的日记内容。此工具支持从数据库拉取并解密原始文字内容。",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"keyword": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词",
				},
				"startTime": map[string]interface{}{
					"type":        "string",
					"description": "开始时间 (格式: yyyy-MM-dd HH:mm:ss)",
				},
				"endTime": map[string]interface{}{
					"type":        "string",
					"description": "结束时间 (格式: yyyy-MM-dd HH:mm:ss)",
				},
			},
			"required": []string{"keyword"},
		},
	}
}

// Execute 真正执行日记搜索请求
func (t *SearchDiaryTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args SearchDiaryArgs) (*mcp.CallToolResult, any, error) {
	apiKey, _ := ctx.Value("apiKey").(string)

	grpcReq := &pb.SearchDiaryRequest{
		ApiKey:    apiKey,
		Keyword:   args.Keyword,
		StartTime: args.StartTime,
		EndTime:   args.EndTime,
	}

	res, err := grpc.SearchDiary(ctx, grpcReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("gRPC error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	var textContent string
	for i, r := range res.Results {
		textContent += fmt.Sprintf("Diary %d [%s] (ID: %s, Emotion: %s):\n%s\n\n",
			i+1, r.Date, r.DiaryId, r.Emotion, r.Content)
	}

	if textContent == "" {
		textContent = "No diaries found matching the criteria."
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textContent},
		},
	}, nil, nil
}

// SearchMemoryArgs 定义了 memorySearch 工具的入参结构
type SearchMemoryArgs struct {
	Query      string `json:"query"`
	MaxResults int32  `json:"maxResults,omitempty"`
}

// SearchMemoryTool 用于执行后端的 SearchMemory gRPC 方法
type SearchMemoryTool struct{}

// NewSearchMemoryTool 创建一个新的 SearchMemoryTool 实例
func NewSearchMemoryTool() *SearchMemoryTool {
	return &SearchMemoryTool{}
}

// GetToolDef 返回该工具在 MCP 中注册的定义
func (t *SearchMemoryTool) GetToolDef() *mcp.Tool {
	return &mcp.Tool{
		Name:        "memorySearch",
		Description: "搜索用户的记忆信息，包括中期记忆（AI 总结的重要事件）和短期记忆上下文（最近的对话记录）。此工具综合了向量检索和对话历史，提供全面的记忆检索能力。",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词或问题",
				},
				"maxResults": map[string]interface{}{
					"type":        "integer",
					"description": "最大返回结果数量（默认 10）",
				},
			},
			"required": []string{"query"},
		},
	}
}

// Execute 真正执行记忆搜索请求
func (t *SearchMemoryTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args SearchMemoryArgs) (*mcp.CallToolResult, any, error) {
	apiKey, _ := ctx.Value("apiKey").(string)

	maxResults := args.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	grpcReq := &pb.SearchMemoryRequest{
		ApiKey:     apiKey,
		Query:      args.Query,
		MaxResults: maxResults,
	}

	res, err := grpc.SearchMemory(ctx, grpcReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("gRPC error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	var textContent string
	for i, r := range res.Results {
		textContent += fmt.Sprintf("Memory %d [Type: %s] (Score: %.2f, Source: %s):\n%s\n\n",
			i+1, r.Type, r.Score, r.SourceId, r.Content)
	}

	if textContent == "" {
		textContent = "No memories found matching the query."
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textContent},
		},
	}, nil, nil
}
