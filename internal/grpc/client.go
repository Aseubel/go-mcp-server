package grpc

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "mcp/proto"
)

var client pb.McpExtensionServiceClient

// InitClient 连接到 Java gRPC 服务端
func InitClient(target string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("无法连接到 gRPC 服务端 %s: %v", target, err)
		return nil, err
	}
	client = pb.NewMcpExtensionServiceClient(conn)
	log.Printf("成功初始化 gRPC 连接，目标地址: %s", target)
	return conn, nil
}

// GetClient 返回已经初始化好的 gRPC 客户端实例
func GetClient() pb.McpExtensionServiceClient {
	if client == nil {
		log.Fatal("gRPC 客户端尚未初始化，请先调用 InitClient。")
	}
	return client
}

// SearchDiary 调用后端的 SearchDiary RPC 接口
func SearchDiary(ctx context.Context, request *pb.SearchDiaryRequest) (*pb.SearchDiaryResponse, error) {
	res, err := GetClient().SearchDiary(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("gRPC SearchDiary failed: %w", err)
	}
	if res.ErrorMessage != "" {
		return nil, fmt.Errorf("后端返回错误: %s", res.ErrorMessage)
	}
	return res, nil
}

// QueryLifeGraph 调用后端的 QueryLifeGraph RPC 接口
func QueryLifeGraph(ctx context.Context, request *pb.QueryLifeGraphRequest) (*pb.QueryLifeGraphResponse, error) {
	res, err := GetClient().QueryLifeGraph(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("gRPC QueryLifeGraph failed: %w", err)
	}
	if res.ErrorMessage != "" {
		return nil, fmt.Errorf("后端返回错误: %s", res.ErrorMessage)
	}
	return res, nil
}
