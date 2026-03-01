@echo off
REM 重新生成 MCP gRPC Proto 文件脚本 (Windows 版本)
REM 使用前请确保已安装以下工具:
REM - protoc (Protocol Buffers Compiler)
REM - protoc-gen-go (Go plugin for protoc)
REM - protoc-gen-go-grpc (Go gRPC plugin for protoc)

echo Installing Go proto plugins...
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

REM 确保插件在 PATH 中
set PATH=%PATH%;%GOPATH%\bin

REM 编译 proto 文件
echo Compiling proto files...
protoc --go_out=. --go_opt=paths=source_relative ^
       --go-grpc_out=. --go-grpc_opt=paths=source_relative ^
       proto\mcp_extension.proto

echo Proto files compiled successfully!
echo Generated files:
echo   - proto\mcp_extension.pb.go
echo   - proto\mcp_extension_grpc.pb.go

pause
