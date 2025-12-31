# MCP Server Integration

This module implements the Model Context Protocol (MCP) Server for the AI Ability service.
It allows the LLM to access external tools and data sources in a standardized way.

## Features

- **Web Search**: Integrated web search using Bing, Serper, or Google.
- **Database Query**: Safe, read-only SQL query execution against the application database.
- **LLM Integration**: Direct integration with OpenAI Chat provider for high-performance tool execution.

## Configuration

Configure the MCP server in your `config.yaml` (or `config_local.yaml`):

```yaml
mcp:
  enabled: true
  search:
    provider: "bing" # bing, serper, google
    api_key: "your-api-key"
    cx: "your-google-cx" # only for google
```

## Architecture

The MCP Server is initialized in `main.go` and stored in the global context.
The HTTP server registers MCP handlers to expose standard MCP endpoints.
The LLM Provider (`LLMOpenAIChatProvider`) retrieves the MCP server from the context and injects registered tools into OpenAI requests.

### Low-Coupling Design

- **Interface-based**: The `SearchTool` uses a `SearchProvider` interface, allowing easy addition of new search engines.
- **Context-aware**: The MCP server is passed via context, decoupling the LLM provider from the MCP implementation details.
- **Standardized Tools**: Tools are converted to standard OpenAI function definitions automatically.

## Usage

### LLM Tool Use

When `mcp.enabled` is true, the OpenAI Chat provider will automatically:
1. Detect registered MCP tools.
2. Add them to the `tools` parameter in the Chat Completion API.
3. Handle tool calls internally:
   - Execute the tool using the MCP server.
   - Feed the result back to the LLM.
   - Continue the conversation until a final response is generated or max turns reached.

### HTTP Endpoints

The MCP server also exposes standard MCP JSON-RPC endpoints (if enabled in routes):
- `/mcp` (WebSocket or HTTP POST depending on transport)

## Extending

To add a new tool:
1. Create a new tool struct in `mcp/tools/`.
2. Implement `GetToolDef()` and `Execute()`.
3. Register the tool in `mcp/server.go`.
