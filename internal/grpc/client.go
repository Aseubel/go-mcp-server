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

// InitClient connects to the Java gRPC server
func InitClient(target string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Did not connect to gRPC server at %s: %v", target, err)
		return nil, err
	}
	client = pb.NewMcpExtensionServiceClient(conn)
	log.Printf("Successfully initialized gRPC client connected to %s", target)
	return conn, nil
}

// GetClient returns the initialized gRPC client
func GetClient() pb.McpExtensionServiceClient {
	if client == nil {
		log.Fatal("gRPC client is not initialized. Call InitClient first.")
	}
	return client
}

// SearchDiary calls the SearchDiary RPC
func SearchDiary(ctx context.Context, request *pb.SearchDiaryRequest) (*pb.SearchDiaryResponse, error) {
	res, err := GetClient().SearchDiary(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("gRPC SearchDiary failed: %w", err)
	}
	if res.ErrorMessage != "" {
		return nil, fmt.Errorf("backend error: %s", res.ErrorMessage)
	}
	return res, nil
}

// QueryLifeGraph calls the QueryLifeGraph RPC
func QueryLifeGraph(ctx context.Context, request *pb.QueryLifeGraphRequest) (*pb.QueryLifeGraphResponse, error) {
	res, err := GetClient().QueryLifeGraph(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("gRPC QueryLifeGraph failed: %w", err)
	}
	if res.ErrorMessage != "" {
		return nil, fmt.Errorf("backend error: %s", res.ErrorMessage)
	}
	return res, nil
}
