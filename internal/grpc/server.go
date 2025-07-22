package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Daylily-kor/daylily-docker-client/internal/docker"
	"github.com/Daylily-kor/daylily-docker-client/internal/logger"
	"github.com/Daylily-kor/daylily-docker-client/proto/dockerpb"
	dockerTypes "github.com/docker/docker/api/types"
)

// Server implements the gRPC DockerService
type Server struct {
	dockerpb.UnimplementedDockerServiceServer
	dockerClient *docker.Client
}

// NewServer creates a new gRPC server
func NewServer(dockerClient *docker.Client) *Server {
	return &Server{
		dockerClient: dockerClient,
	}
}

// Version returns the Docker server version
func (s *Server) Version(ctx context.Context, in *emptypb.Empty) (*dockerpb.VersionResponse, error) {
	dockerVersion, err := s.dockerClient.GetVersion(ctx)
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
func (s *Server) Build(ctx context.Context, in *dockerpb.BuildRequest) (*dockerpb.BuildResponse, error) {
	logger.Info("Received Build request",
		"owner", in.Owner,
		"repository", in.Repository,
		"ref", in.Ref)

	req := &dockerpb.BuildRequest{
		Owner:      in.Owner,
		Repository: in.Repository,
		Ref:        in.Ref,
	}

	resp, err := s.dockerClient.Build(ctx, req)
	if err != nil {
		logger.Error("Error building image", "error", logger.WithError(err))
		return nil, status.Errorf(codes.Internal, "failed to build image: %v", err)
	}

	logger.Info("Successfully built image", "image_name", resp.ImageName)

	return &dockerpb.BuildResponse{ImageName: resp.ImageName}, nil
}

// Run starts a Docker container from an image
func (s *Server) Run(ctx context.Context, in *dockerpb.RunRequest) (*dockerpb.RunResponse, error) {
	req := &dockerpb.RunRequest{
		ImageName: in.ImageName,
	}

	resp, err := s.dockerClient.Run(ctx, req)
	if err != nil {
		logger.Error("Error running container",
			"image_name", in.ImageName,
			"error", logger.WithError(err))
		return nil, status.Errorf(codes.Internal, "failed to run container: %v", err)
	}

	return &dockerpb.RunResponse{
		ContainerId: resp.ContainerId,
		Status:      resp.Status,
	}, nil
}

// RegisterServer registers the DockerService server with the gRPC server
func RegisterServer(grpcServer *grpc.Server, dockerClient *docker.Client) {
	server := NewServer(dockerClient)
	dockerpb.RegisterDockerServiceServer(grpcServer, server)
}

// toVersionResponse converts Docker version to protobuf response
func toVersionResponse(version *dockerTypes.Version) *dockerpb.VersionResponse {
	return &dockerpb.VersionResponse{
		Version:       version.Version,
		ApiVersion:    version.APIVersion,
		Platform:      version.Platform.Name,
		Os:            version.Os,
		Arch:          version.Arch,
		KernelVersion: version.KernelVersion,
	}
}
