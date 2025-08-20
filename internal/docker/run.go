package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/Daylily-kor/daylily-grpc-server/internal/logger"
	"github.com/Daylily-kor/daylily-grpc-server/pb/run"

	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerNetwork "github.com/docker/docker/api/types/network"
)

// Run starts a Docker container from an image
func (c *Client) Run(ctx context.Context, req *run.GrpcContainerRunRequest) (*run.GrpcContainerRunResponse, error) {
	// Find the Docker network used by Traefik
	networkName, err := c.FindTraefikNetwork(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find Traefik network: %w", err)
	} else {
		logger.Debug("Found Traefik network", "network_name", networkName)
	}

	// Discover the exposed port for the image and use the first one
	port, err := c.DiscoverPorts(ctx, req.ImageId)
	if err != nil {
		return nil, fmt.Errorf("failed to discover ports for image %s: %w", req.ImageId, err)
	} else {
		logger.Debug("Discovered port for image", "port", port, "image_id", req.ImageId)
	}

	// Configure the network for the container
	networkConfig := &dockerNetwork.NetworkingConfig{
		EndpointsConfig: map[string]*dockerNetwork.EndpointSettings{
			networkName: {NetworkID: networkName}, // Attach to the Traefik network
		},
	}

	containerName := req.ContainerName
	containerUrl := fmt.Sprintf("%s.%s", containerName, req.BaseDomain)
	containerConfig := &dockerContainer.Config{
		Image: req.ImageId,
		Tty:   false,
		Labels: map[string]string{
			"traefik.enable": "true",
			"traefik.http.routers." + containerName + ".rule":                      "Host(`" + containerUrl + "`)",
			"traefik.http.services." + containerName + ".loadbalancer.server.port": port,
			"daylily.container":           "true",
			"daylily.container.commitSHA": req.CommitSHA,
		},
	}

	listOptions := dockerContainer.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "name", Value: containerName},
		),
	}

	containers, err := c.ContainerList(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	logger.Debug("Found existing containers", "count", len(containers), "name", containerName)

	// If the container already exists, remove it
	if len(containers) > 0 {
		if err := c.ContainerStop(ctx, containers[0].ID, dockerContainer.StopOptions{}); err != nil {
			return nil, fmt.Errorf("failed to stop existing container: %w", err)
		}

		logger.Debug("Removing existing container", "container_id", containers[0].ID, "name", containerName, "state", containers[0].State)

		if err := c.ContainerRemove(ctx, containers[0].ID, dockerContainer.RemoveOptions{}); err != nil {
			return nil, fmt.Errorf("failed to remove existing container: %w", err)
		}
	}

	// Create the container
	createResp, err := c.ContainerCreate(ctx, containerConfig, nil, networkConfig, nil, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	} else {
		logger.Debug("Container created", "container_id", createResp.ID, "name", containerName)
	}

	// Start the container
	if err := c.ContainerStart(ctx, createResp.ID, dockerContainer.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Inspect the container to check its status
	inspectResp, err := c.ContainerInspect(ctx, createResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	} else {
		logger.Debug("Container inspected", "container_id", createResp.ID, "status", inspectResp.State.Status)
	}

	// Fire container stop after 1 minute
	go func(id string) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		<-ctx.Done()
		logger.Debug("Stopping container after timeout", "container_id", id)

		if err := c.ContainerStop(context.Background(), id, dockerContainer.StopOptions{}); err != nil {
			logger.Error("Failed to stop container", "container_id", id, "error", err)
		} else {
			logger.Info("Container stopped successfully", "container_id", id)
		}
	}(createResp.ID)

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

	return &run.GrpcContainerRunResponse{
		ContainerId:   createResp.ID,
		ContainerName: containerName,
		Status:        inspectResp.State.Status,
	}, nil
}
