package docker

import (
	"context"
	"fmt"

	"github.com/Daylily-kor/daylily-docker-client/internal/logger"
	"github.com/Daylily-kor/daylily-docker-client/proto/dockerpb"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerNetwork "github.com/docker/docker/api/types/network"
	"github.com/lithammer/shortuuid/v4"
)

// Run starts a Docker container from an image
func (c *Client) Run(ctx context.Context, req *dockerpb.RunRequest) (*dockerpb.RunResponse, error) {
	networkName, err := c.FindTraefikNetwork(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find Traefik network: %w", err)
	} else {
		logger.Debug("Found Traefik network", "network_name", networkName)
	}

	port, err := c.DiscoverPorts(ctx, req.ImageName)
	if err != nil {
		return nil, fmt.Errorf("failed to discover ports for image %s: %w", req.ImageName, err)
	} else {
		logger.Debug("Discovered port for image", "port", port, "image", req.ImageName)
	}

	// Configure the network for the container
	networkConfig := &dockerNetwork.NetworkingConfig{
		EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
			networkName: {NetworkID: networkName}, // Attach to the Traefik network
		},
	}

	containerName := shortuuid.New()
	createResp, err := c.ContainerCreate(
		ctx,
		&dockerContainer.Config{
			Image: req.ImageName,
			Tty:   false,
			Labels: map[string]string{
				"traefik.enable": "true",
				"traefik.http.routers." + containerName + ".rule":                      "Host(`test.docker.localhost`)",
				"traefik.http.services." + containerName + ".loadbalancer.server.port": port,
			},
		},
		nil,
		networkConfig,
		nil,
		containerName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	} else {
		logger.Debug("Container created", "container_id", createResp.ID, "name", containerName)
	}

	if err := c.ContainerStart(ctx, createResp.ID, dockerContainer.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	inspectResp, err := c.ContainerInspect(ctx, createResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	} else {
		logger.Debug("Container inspected", "container_id", createResp.ID, "status", inspectResp.State.Status)
	}

	switch inspectResp.State.Status {
	case "created", "restarting", "running":
		logger.Info("Container is running",
			"container_id", createResp.ID,
			"status", inspectResp.State.Status)
	case "exited":
		return nil, fmt.Errorf("container %s has exited", createResp.ID)
	default:
		return nil, fmt.Errorf("container %s is in an unknown state: %s", createResp.ID, inspectResp.State.Status)
	}

	return &dockerpb.RunResponse{
		ContainerId: createResp.ID,
		Status:      "Container started successfully",
	}, nil
}
