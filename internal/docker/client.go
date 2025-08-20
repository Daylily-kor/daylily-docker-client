package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/Daylily-kor/daylily-grpc-server/internal/logger"
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

// FindTraefikNetwork finds the network used by the Traefik container
func (c *Client) FindTraefikNetwork(ctx context.Context) (string, error) {
	// List all containers with the label
	args := filters.NewArgs()
	args.Add("label", "com.docker.compose.service=traefik")
	containers, err := c.ContainerList(ctx, dockerContainer.ListOptions{All: true, Filters: args})
	if err != nil || len(containers) == 0 {
		return "", fmt.Errorf("failed to find traefik container: %w", err)
	}

	// Get the network name from the traefik container
	for networkName := range containers[0].NetworkSettings.Networks {
		return networkName, nil
	}

	// Probably should not reach
	return "", fmt.Errorf("traefik has no networks")
}

// DiscoverPorts discovers the exposed ports of a Docker image
func (c *Client) DiscoverPorts(ctx context.Context, imageID string) (string, error) {
	// Inspect the docker image of given Image ID
	inspectResp, err := c.ImageInspect(ctx, imageID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect image %s: %w", imageID, err)
	}

	if len(inspectResp.Config.ExposedPorts) == 0 {
		return "", fmt.Errorf("no exposed ports found for image %s", imageID)
	}

	// Return the first exposed port found
	for port := range inspectResp.Config.ExposedPorts {
		logger.Debug("Discovered exposed port", "port", port)
		// ex: 80/tcp, 12345/tcp -> needs to be sliced and return only port number
		portParts := strings.Split(port, "/")
		if len(portParts) > 0 {
			return portParts[0], nil
		}
	}

	return "", fmt.Errorf("no valid exposed ports found for image %s", imageID)
}
