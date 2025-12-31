package tools

import (
	"context"
	"fmt"
	"strings"

	"mcp/config"
	search_utils "mcp/tools/search"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchArgs defines the arguments for the search tool
type SearchArgs struct {
	Query string `json:"query" jsonschema:"required"`
}

// SearchProvider interface for different search engines
type SearchProvider interface {
	Search(ctx context.Context, query string) (string, error)
}

// SearchTool implements the web search tool
type SearchTool struct {
	Config   config.SearchConfig
	Provider SearchProvider
}

// NewSearchTool creates a new search tool
func NewSearchTool(cfg config.SearchConfig) *SearchTool {
	var provider SearchProvider
	switch strings.ToLower(cfg.Provider) {
	case "bing":
		provider = &search_utils.BingProvider{APIKey: cfg.APIKey}
	case "serper":
		provider = &search_utils.SerperProvider{APIKey: cfg.APIKey}
	case "bocha":
		provider = &search_utils.BochaProvider{APIKey: cfg.APIKey}
	default:
		// Default to Google
		provider = &search_utils.GoogleProvider{APIKey: cfg.APIKey, CX: cfg.CX}
	}
	return &SearchTool{Config: cfg, Provider: provider}
}

// GetToolDef returns the MCP tool definition
func (t *SearchTool) GetToolDef() *mcp.Tool {
	desc := fmt.Sprintf("Search the internet for real-time information using %s. Use this tool when you need to find current events, facts, or information not present in your training data.", t.Config.Provider)
	if t.Config.Provider == "" {
		desc = "Search the internet for real-time information using Google Search. Use this tool when you need to find current events, facts, or information not present in your training data."
	}

	return &mcp.Tool{
		Name:        "web_search",
		Description: desc,
		InputSchema:  t.getInputSchema(),
	}
}

func (t *SearchTool) getInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
		},
		"required": []string{"query"},
	}
}

// Register registers the tool with the MCP server
// Deprecated: Use GetToolDef and server.RegisterTool instead
func (t *SearchTool) Register(s *mcp.Server) {
	mcp.AddTool(s, t.GetToolDef(), t.Execute)
}

// Execute performs the search
func (t *SearchTool) Execute(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	resultText, err := t.Provider.Search(ctx, args.Query)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error performing search: %v", err)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultText},
		},
	}, nil, nil
}
