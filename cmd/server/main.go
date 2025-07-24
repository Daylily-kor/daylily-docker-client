package main

import (
	"net"

	"google.golang.org/grpc"

	"github.com/Daylily-kor/daylily-grpc-server/internal/docker"
	grpcServer "github.com/Daylily-kor/daylily-grpc-server/internal/grpc"
	"github.com/Daylily-kor/daylily-grpc-server/internal/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	// Create Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		logger.Fatal("Failed to create Docker client", "error", err)
	}
	defer dockerClient.Close()

	// Create gRPC server with interceptors
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcServer.LoggingInterceptor),
		grpc.ChainUnaryInterceptor(grpcServer.RecoveryInterceptor),
	)

	// Register the Docker service
	grpcServer.RegisterServer(server, dockerClient)

	// Create listener
	// TODO: 서버 주소와 포트를 구성파일로 설정할 수 있도록 변경
	listener, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		logger.Fatal("Failed to listen on port 50051", "error", err)
	}

	logger.Info("Starting gRPC server", "port", "50051")
	if err := server.Serve(listener); err != nil {
		logger.Fatal("Failed to serve gRPC server", "error", err)
	}
}
