# MCP Server 集成 (Model Context Protocol)

本模块实现了针对 AI 能力服务的 Model Context Protocol (MCP) Server。
它允许大语言模型 (LLM) 以标准化的方式访问外部工具和数据源。

## 功能特性

- **网络搜索**: 集成了基于 Bocha、Serper 或 Google 的网络实时搜索功能。
- **与 Java 后端集成**: 通过 gRPC 连接到基础应用后端，实现双向数据交互和功能级扩展。
- **特定工具扩展**:
  - `diarySearch`: 根据关键词和可选的时间范围查询用户的日记内容。
  - `lifeGraph`: 查询用户的生命图谱（时空关系知识库）以获取人物、事件的上下文关系。

## 配置说明

在你的 `config.yaml` 或系统环境变量中配置 MCP 服务：

```yaml
server:
  port: 11611
  env: "dev"

search:
  provider: "bocha" # 支持 bocha, serper, google 等
  api_key: "your-api-key"
  cx: "your-google-cx" # 仅用于 google 搜索

grpc:
  backend_target: "localhost:9090" # Java 后端 gRPC 地址

log:
  level: "debug"
```

## 架构设计

MCP Server 的核心在 `server.go` 中初始化。
整个系统采用低耦合、易扩展的设计：

- **基于接口的搜索机制**: `SearchTool` 内部定义了 `Provider` 接口，可以非常轻松地添加新的搜索引擎而不影响外部逻辑。
- **无侵入的工具注册**: 工具的定义 (`GetToolDef`) 与执行 (`Execute`) 被解耦到专门的结构体中（见 `internal/tools` 和 `tools` 目录），并提供统一的注册口。
- **与 Java 系统的无缝互调**: 基于 Protobuf / gRPC 进行跨语言通讯，可以直接查询、调用远端的业务核心域。

## 服务端点 (Endpoints)

MCP Server 目前支持以下访问方式：

- **Streamable HTTP**: `POST /mcp` （推荐使用）
- **传统 SSE 机制**: `GET /sse` 与 `POST /messages`

可以在主应用或其他客户端中直接配置该 MCP 服务的访问 URL 进行调用。

## 扩展与使用指南

想要添加新的能力 / 工具？请遵循以下步骤：

1. 在 `mcp/internal/tools` 或 `mcp/tools` 下创建一个新的工具结构体定义。
2. 实现该工具的两个核心方法：
   - `GetToolDef() *mcp.Tool`: 定义工具的名称、描述以及 JSON Schema 入参结构。
   - `Execute(ctx, req, args) (*mcp.CallToolResult, any, error)`: 实现工具请求的具体处理逻辑。
3. 在 `mcp/server.go` 的 `NewMCPServer` 函数中，使用 `RegisterTool` 进行工具注册。
