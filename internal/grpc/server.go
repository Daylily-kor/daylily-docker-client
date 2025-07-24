package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Daylily-kor/daylily-grpc-server/internal/docker"
	"github.com/Daylily-kor/daylily-grpc-server/internal/logger"
	"github.com/Daylily-kor/daylily-grpc-server/pb/build"
	"github.com/Daylily-kor/daylily-grpc-server/pb/run"
	"github.com/Daylily-kor/daylily-grpc-server/pb/service"
	"github.com/Daylily-kor/daylily-grpc-server/pb/version"
	dockerTypes "github.com/docker/docker/api/types"
)

// Server implements the gRPC DockerService
type Server struct {
	service.UnimplementedDockerServiceServer
	dockerClient *docker.Client
}

// NewServer creates a new gRPC server
func NewServer(dockerClient *docker.Client) *Server {
	return &Server{
		dockerClient: dockerClient,
	}
}

// Version returns the Docker server version
func (s *Server) Version(ctx context.Context, _ *emptypb.Empty) (*version.VersionResponse, error) {
	dockerVersion, err := s.dockerClient.Version(ctx)
	if err != nil {
		logger.Error("Error getting Docker server version", "error", logger.WithError(err))
		return nil, status.Errorf(codes.Internal, "failed to get Docker server version: %v", err)
	}

	logger.Info("Retrieved Docker server version",
		"version", dockerVersion.Version,
		"api_version", dockerVersion.APIVersion)

	return toVersionResponse(dockerVersion), nil
}

// Build builds a Docker image from a GitHub repository
func (s *Server) Build(ctx context.Context, in *build.ImageBuildRequest) (*build.ImageBuildResponse, error) {
	resp, err := s.dockerClient.Build(ctx, in)
	if err != nil {
		logger.Error("Error building image", "error", logger.WithError(err))
		return nil, status.Errorf(codes.Internal, "failed to build image: %v", err)
	}

	logger.Info("Successfully built image",
		"image_name", resp.ImageName,
		"image_id", resp.ImageId)
	return resp, nil
}

// Run starts a Docker container from an image
func (s *Server) Run(ctx context.Context, in *run.RunRequest) (*run.RunResponse, error) {
	resp, err := s.dockerClient.Run(ctx, in)
	if err != nil {
		logger.Error("Error running container",
			"image_id", in.ImageId,
			"error", logger.WithError(err))
		return nil, status.Errorf(codes.Internal, "failed to run container: %v", err)
	}

	logger.Info("Successfully started container",
		"container_id", resp.ContainerId,
		"status", resp.Status)

	return resp, nil
}

// RegisterServer registers the DockerService server with the gRPC server
func RegisterServer(grpcServer *grpc.Server, dockerClient *docker.Client) {
	server := NewServer(dockerClient)
	service.RegisterDockerServiceServer(grpcServer, server)
}

// toVersionResponse converts Docker version to protobuf response
func toVersionResponse(v *dockerTypes.Version) *version.VersionResponse {
	return &version.VersionResponse{
		Version:       v.Version,
		ApiVersion:    v.APIVersion,
		Platform:      v.Platform.Name,
		Os:            v.Os,
		Arch:          v.Arch,
		KernelVersion: v.KernelVersion,
	}
}
