package docker

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"

	"github.com/Daylily-kor/daylily-grpc-server/internal/logger"
	"github.com/Daylily-kor/daylily-grpc-server/pb/build"
	"github.com/Daylily-kor/daylily-grpc-server/util"
	dockerBuildTypes "github.com/docker/docker/api/types/build"
	"github.com/moby/buildkit/session"
)

// Build builds a Docker image from a GitHub repository
func (c *Client) Build(ctx context.Context, req *build.ImageBuildRequest) (*build.ImageBuildResponse, error) {
	// https://github.com/moby/moby/issues/48112#issuecomment-2916141864
	sess, err := session.NewSession(ctx, "secret123")
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	go func() {
		if err := sess.Run(ctx, func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
			return c.DialHijack(ctx, "/session", proto, meta)
		}); err != nil {
			logger.Error("Error running session", "error", err)
		}
	}()
	defer sess.Close()

	// URL to the remote GitHub repository
	// Ex: https://github.com/Daylily-kor/daylily-grpc-server.git#main
	remote := fmt.Sprintf("https://github.com/%s.git#%s", req.RepositoryName, req.Ref)

	// Tags for the image
	// Ex: pr-<pr_number>-<commit_hash>
	tag := fmt.Sprintf("pr-%d-%s", req.PrNumber, req.Sha)

	// Docker image name
	// Ex: Daylily-kor/daylily-grpc-server:<tag>
	imageName := req.RepositoryName + ":" + tag

	opts := dockerBuildTypes.ImageBuildOptions{
		RemoteContext: remote,
		Tags:          []string{imageName},
		Version:       "2",
		SessionID:     sess.ID(),
	}

	// Build image
	imageBuildResp, err := c.ImageBuild(ctx, bytes.NewReader(nil), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}
	defer imageBuildResp.Body.Close()

	// Stream build output to stdout
	if err := util.StreamBuildOutput(ctx, imageBuildResp.Body, os.Stdout); err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	imageInspectResp, err := c.ImageInspect(ctx, imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %w", err)
	}

	return &build.ImageBuildResponse{
		ImageId:   imageInspectResp.ID,
		ImageName: req.RepositoryName,
		ImageTag:  tag,
	}, nil
}
