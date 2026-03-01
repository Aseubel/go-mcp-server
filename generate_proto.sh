#!/bin/bash

# 重新生成 MCP gRPC Proto 文件脚本
# 使用前请确保已安装以下工具：
# - protoc (Protocol Buffers Compiler)
# - protoc-gen-go (Go plugin for protoc)
# - protoc-gen-go-grpc (Go gRPC plugin for protoc)

# 安装 protoc-gen-go 和 protoc-gen-go-grpc
echo "Installing Go proto plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 确保插件在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin

# 编译 proto 文件
echo "Compiling proto files..."
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/mcp_extension.proto

echo "Proto files compiled successfully!"
echo "Generated files:"
echo "  - proto/mcp_extension.pb.go"
echo "  - proto/mcp_extension_grpc.pb.go"
