package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
)

// Client wraps the Docker client with additional functionality
type Client struct {
	*dockerClient.Client
}

// NewClient creates a new Docker client wrapper
func NewClient() (*Client, error) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &Client{Client: client}, nil
}

// GetVersion returns the Docker server version
func (c *Client) GetVersion(ctx context.Context) (*types.Version, error) {
	version, err := c.ServerVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker server version: %w", err)
	}
	return &version, nil
}

// FindTraefikNetwork finds the network used by the Traefik container
func (c *Client) FindTraefikNetwork(ctx context.Context) (string, error) {
	args := filters.NewArgs()
	// Should be "daylily" in production
	args.Add("label", "com.docker.compose.service=traefik")
	ctrs, err := c.ContainerList(ctx, dockerContainer.ListOptions{All: true, Filters: args})
	if err != nil || len(ctrs) == 0 {
		return "", fmt.Errorf("failed to find traefik container: %w", err)
	}

	for networkName := range ctrs[0].NetworkSettings.Networks {
		return networkName, nil
	}

	return "", fmt.Errorf("traefik has no networks")
}

// DiscoverPorts discovers the exposed ports of a Docker image
func (c *Client) DiscoverPorts(ctx context.Context, imageID string) (string, error) {
	inspectResp, err := c.ImageInspect(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect image %s: %w", imageID, err)
	}

	if len(inspectResp.Config.ExposedPorts) == 0 {
		return "", fmt.Errorf("no exposed ports found for image %s", imageID)
	}

	for port := range inspectResp.Config.ExposedPorts {
		return port, nil
	}

	return "", fmt.Errorf("no valid exposed ports found for image %s", imageID)
}
