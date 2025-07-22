package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Daylily-kor/daylily-docker-client/internal/logger"
)

// LoggingInterceptor logs gRPC requests
func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	logger.Info("gRPC request started", "method", info.FullMethod)

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		logger.Error("gRPC request failed",
			"method", info.FullMethod,
			"duration", logger.WithDuration(duration),
			"error", logger.WithError(err))
	} else {
		logger.Info("gRPC request completed",
			"method", info.FullMethod,
			"duration", logger.WithDuration(duration))
	}

	return resp, err
}

// RecoveryInterceptor recovers from panics in gRPC handlers
func RecoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in gRPC handler",
				"method", info.FullMethod,
				"panic", r)
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
