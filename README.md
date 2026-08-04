# MCP Server 集成 (Model Context Protocol)

本模块实现了针对 AI 能力服务的 Model Context Protocol (MCP) Server。
它允许大语言模型 (LLM) 以标准化的方式访问外部工具和数据源。

## 功能特性

- **网络搜索**: 集成了基于 Bocha、Serper 或 Google 的网络实时搜索功能。
- **与 Java 后端集成**: Go MCP Server 通过 gRPC 调用 Java 后端的内部记忆函数；Java gRPC 边界负责 API Key 和 scope 鉴权。
- **特定工具扩展**:
  - `diarySearch`: 根据关键词和可选的时间范围查询用户的日记内容。
  - `lifeGraph`: 查询用户的生命图谱（时空关系知识库）以获取人物、事件的上下文关系。

## 配置说明

在你的 `config.yaml` 或系统环境变量中配置 MCP 服务：

```yaml
server:
  port: 11611
  env: "dev"
  auth_api_key: "set-a-long-random-internal-service-key"
  
search:
  provider: "bocha" # 支持 bocha, serper, google 等
  api_key: "your-api-key"
  cx: "your-google-cx" # 仅用于 google 搜索

grpc:
  backend_target: "localhost:9090" # Java 后端 gRPC 地址

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

同一个 Go MCP 进程提供两个入口，工具列表和认证方式不同：

- 公共入口 `/mcp`：
  - **diarySearch**: 根据关键词和可选的时间范围搜索当前用户的日记内容
  - **memorySearch**: 搜索当前用户的中期记忆和短期记忆上下文
- 内部入口 `/internal/mcp`：
  - **web_search**: 供 Java 后端使用的网络搜索工具

公共入口使用用户 Developer API Key；内部入口使用 `X-MCP-Service-Key`。公共入口不会返回或执行 `web_search`，内部入口不会返回或执行记忆工具。

## 客户端配置示例

### Claude Desktop (MacOS / Windows)

外部客户端应通过反向代理发布的 HTTPS 公共入口接入；推荐使用 Streamable HTTP。

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
         "url": "https://mcp.example.com/mcp",
         "headers": {
           "Authorization": "Bearer <developer-key>"
         }
       }
     }
   }
   ```

   *注意：公共入口只接受用户 Developer API Key，可通过 `Authorization: Bearer <developer-key>`、`X-Developer-API-Key` 或 `X-API-Key` 传递。`X-MCP-Service-Key` 只用于 Java 到 `/internal/mcp` 的内部调用，不应发给外部用户。*

## 架构设计

MCP Server 的核心在 `server.go` 中初始化。
整个系统采用低耦合、易扩展的设计：

- **基于接口的搜索机制**: `SearchTool` 内部定义了 `Provider` 接口，可以非常轻松地添加新的搜索引擎而不影响外部逻辑。
- **无侵入的工具注册**: 工具的定义 (`GetToolDef`) 与执行 (`Execute`) 被解耦到专门的结构体中（见 `internal/tools` 和 `tools` 目录），并提供统一的注册口。
- **Java 内部能力复用**: 基于 Protobuf / gRPC 调用 Java 内部函数；Go 工具不复制记忆查询、解密和权限逻辑。

## 服务端点 (Endpoints)

公共端点：

- **Streamable HTTP**: `POST /mcp` （推荐使用）
- **传统 SSE 机制**: `GET /sse` 与 `POST /messages`
- 工具：`diarySearch`、`memorySearch`
- 认证：Developer API Key

内部端点：

- **Streamable HTTP**: `POST /internal/mcp` （Java 后端使用）
- **传统 SSE 机制**: `GET /internal/sse` 与 `POST /internal/messages`
- 工具：`web_search`
- 认证：`X-MCP-Service-Key`

公共 MCP 不要求来源登记或 CORS 白名单；用户 Agent 直接通过 HTTPS 和 Developer API Key 连接即可。若未来需要浏览器页面直连，应在反向代理层单独配置 CORS，不要把它当作 MCP 认证。生产环境应通过 HTTPS 反向代理发布公共 `/mcp`，不要公开 Java gRPC `9090`。

## 扩展与使用指南

想要添加新的能力 / 工具？请遵循以下步骤：

1. 在 `mcp/internal/tools` 或 `mcp/tools` 下创建一个新的工具结构体定义。
2. 实现该工具的两个核心方法：
   - `GetToolDef() *mcp.Tool`: 定义工具的名称、描述以及 JSON Schema 入参结构。
   - `Execute(ctx, req, args) (*mcp.CallToolResult, any, error)`: 实现工具请求的具体处理逻辑。
3. 在 `mcp/server.go` 的 `NewMCPServer` 函数中，使用 `RegisterTool` 进行工具注册。
