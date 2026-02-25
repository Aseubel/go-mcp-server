package tools

import (
	"context"
	"fmt"

	"mcp/internal/grpc"
	pb "mcp/proto"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchDiaryArgs represents the arguments for the diarySearch tool
type SearchDiaryArgs struct {
	UserID    string `json:"userId"`
	Keyword   string `json:"keyword"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

// SearchDiaryTool executes the gRPC SearchDiary method
type SearchDiaryTool struct{}

// NewSearchDiaryTool creates a new SearchDiaryTool
func NewSearchDiaryTool() *SearchDiaryTool {
	return &SearchDiaryTool{}
}

// GetToolDef returns the MCP tool definition
func (t *SearchDiaryTool) GetToolDef() *mcp.Tool {
	return &mcp.Tool{
		Name:        "diarySearch",
		Description: "Search user diaries by keyword and optional time range. Important: Ensure correct user ID is passed contextually. Supports fetching decrypted raw text from DB.",
	}
}

// Execute performs the diary search
func (t *SearchDiaryTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args SearchDiaryArgs) (*mcp.CallToolResult, any, error) {
	grpcReq := &pb.SearchDiaryRequest{
		UserId:    args.UserID,
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

// QueryLifeGraphArgs represents the arguments for the lifeGraph tool
type QueryLifeGraphArgs struct {
	UserID string `json:"userId"`
	Query  string `json:"query"`
}

// QueryLifeGraphTool executes the gRPC QueryLifeGraph method
type QueryLifeGraphTool struct{}

// NewQueryLifeGraphTool creates a new QueryLifeGraphTool
func NewQueryLifeGraphTool() *QueryLifeGraphTool {
	return &QueryLifeGraphTool{}
}

// GetToolDef returns the MCP tool definition
func (t *QueryLifeGraphTool) GetToolDef() *mcp.Tool {
	return &mcp.Tool{
		Name:        "lifeGraph",
		Description: "Queries the user's life graph (spatio-temporal relationship knowledge base) for context on relationships, events, and nodes based on a query.",
	}
}

// Execute performs the life graph query
func (t *QueryLifeGraphTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args QueryLifeGraphArgs) (*mcp.CallToolResult, any, error) {
	grpcReq := &pb.QueryLifeGraphRequest{
		UserId: args.UserID,
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
