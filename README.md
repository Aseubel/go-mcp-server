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
  api_key: "sk-ys-XXX"             # 从控制台个人设置页面生成的 API Key

log:
  level: "debug"
```

## Proto 文件重新生成

当修改了 `proto/mcp_extension.proto` 文件后，需要重新生成 Go 的 proto 代码：

### Windows 系统

运行批处理脚本：
```bash
generate_proto.bat
```

### Linux/Mac 系统

运行 Shell 脚本：
```bash
chmod +x generate_proto.sh
./generate_proto.sh
```

### 手动安装依赖

如果脚本执行失败，请手动安装以下工具：

1. 安装 protoc（Protocol Buffers 编译器）
   - Windows: 从 https://github.com/protocolbuffers/protobuf/releases 下载
   - Linux: `apt-get install protobuf-compiler` 或 `yum install protobuf-compiler`
   - Mac: `brew install protobuf`

2. 安装 Go 插件：
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

3. 确保插件在 PATH 中：
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

4. 编译 proto 文件：
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/mcp_extension.proto
```

## 工具列表

当前 MCP 服务提供以下工具：

- **diarySearch**: 根据关键词和可选的时间范围搜索用户的日记内容
- **memorySearch**: 搜索用户的记忆信息，包括中期记忆（AI 总结的重要事件）和短期记忆上下文（最近的对话记录）

## 客户端配置示例

### Claude Desktop (MacOS / Windows)

本服务支持通过 SSE (Server-Sent Events) 协议接入 Claude Desktop。

1. **启动服务**
   
   确保本服务已在本地启动（默认端口 11611）：
   ```bash
   go run cmd/server/main.go
   # 或者运行编译后的二进制文件
   ```

2. **配置 Claude Desktop**

   编辑配置文件：
   - MacOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`

   在 `mcpServers` 节点下添加配置：

   ```json
   {
     "mcpServers": {
       "yusi-mcp": {
         "url": "http://localhost:11611/sse?api_key=your-secret-key"
       }
     }
   }
   ```

   *注意：如果需要鉴权，必须在 URL 中携带 api_key（推荐）或者确保客户端支持通过 Authorization 头传递。该 Key 会被透传至后端服务进行验证。*

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
