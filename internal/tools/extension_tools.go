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

// QueryLifeGraphArgs 定义了 lifeGraph 工具的入参结构
type QueryLifeGraphArgs struct {
	Query string `json:"query"`
}

// QueryLifeGraphTool 用于执行后端的 QueryLifeGraph gRPC 方法
type QueryLifeGraphTool struct{}

// NewQueryLifeGraphTool 创建一个新的 QueryLifeGraphTool 实例
func NewQueryLifeGraphTool() *QueryLifeGraphTool {
	return &QueryLifeGraphTool{}
}

// GetToolDef 返回该工具在 MCP 中注册的定义
func (t *QueryLifeGraphTool) GetToolDef() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lifeGraph",
		Description: "查询用户的生命图谱（时空关系知识库），基于你的查询问题，获取关于人际关系、事件和重要实体的长程记忆上下文信息。",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "查询关键词",
				},
			},
			"required": []string{"query"},
		},
	}
}

// Execute 真正执行生命图谱查询请求
func (t *QueryLifeGraphTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args QueryLifeGraphArgs) (*mcp.CallToolResult, any, error) {
	apiKey, _ := ctx.Value("apiKey").(string)

	grpcReq := &pb.QueryLifeGraphRequest{
		ApiKey: apiKey,
		Query:  args.Query,
	}

	res, err := grpc.QueryLifeGraph(ctx, grpcReq)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("gRPC error: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	textContent := res.ResultJson
	if textContent == "" {
		textContent = "No relevant context found in the life graph."
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: textContent},
		},
	}, nil, nil
}
