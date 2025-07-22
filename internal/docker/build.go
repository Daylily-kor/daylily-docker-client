package docker

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"

	"github.com/Daylily-kor/daylily-docker-client/internal/logger"
	"github.com/Daylily-kor/daylily-docker-client/proto/dockerpb"
	"github.com/Daylily-kor/daylily-docker-client/util"
	"github.com/docker/docker/api/types/build"
	"github.com/moby/buildkit/session"
)

// Build builds a Docker image from a GitHub repository
func (c *Client) Build(ctx context.Context, req *dockerpb.BuildRequest) (*dockerpb.BuildResponse, error) {
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

	remote := fmt.Sprintf("https://github.com/%s/%s.git#%s", req.Owner, req.Repository, req.Ref)
	pr, sha, err := util.RandomBuildImageName()
	if err != nil {
		return nil, fmt.Errorf("failed to generate random image name: %w", err)
	}

	imageName := fmt.Sprintf("%s/%s", req.Owner, req.Repository)
	tags := []string{
		fmt.Sprintf("%s:pr-%d-%s", imageName, pr, sha),
	}

	imageBuildResp, err := c.ImageBuild(ctx, bytes.NewReader(nil), build.ImageBuildOptions{
		RemoteContext: remote,
		Tags:          tags,
		Version:       "2",
		SessionID:     sess.ID(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build image: %w", err)
	}
	defer imageBuildResp.Body.Close()

	if err := util.StreamBuildOutput(ctx, imageBuildResp.Body, os.Stdout); err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	return &dockerpb.BuildResponse{ImageName: imageName}, nil
}
