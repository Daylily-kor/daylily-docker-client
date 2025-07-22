package main

import (
	"net"

	"google.golang.org/grpc"

	"github.com/Daylily-kor/daylily-docker-client/internal/docker"
	grpcServer "github.com/Daylily-kor/daylily-docker-client/internal/grpc"
	"github.com/Daylily-kor/daylily-docker-client/internal/logger"
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
	listener, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		logger.Fatal("Failed to listen on port 50051", "error", err)
	}

	logger.Info("Starting gRPC server", "port", "50051")
	if err := server.Serve(listener); err != nil {
		logger.Fatal("Failed to serve gRPC server", "error", err)
	}
}
