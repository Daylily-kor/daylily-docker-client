package docker

import (
	"context"
	"strings"

	"github.com/Daylily-kor/daylily-grpc-server/internal/logger"
	"github.com/Daylily-kor/daylily-grpc-server/pb/containerList"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

func (c *Client) ListContainers(ctx context.Context) ([]*containerList.GrpcContainerResponse, error) {
	listOptions := dockerContainer.ListOptions{
		All: false,
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: "daylily.container=true"},
		),
	}

	containers, err := c.ContainerList(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	logger.Debug("Retrieved list of containers",
		"container_count", len(containers))

	response := make([]*containerList.GrpcContainerResponse, 0, len(containers))
	for _, container := range containers {
		// Remove '/' prefix to get container name
		name := strings.Split(container.Names[0], "/")[1]
		r := &containerList.GrpcContainerResponse{
			Id:        container.ID,
			Name:      name,
			Url:       name, // Assuming the name is used as the URL
			State:     container.State,
			Status:    container.Status,
			CommitSHA: container.Labels["daylily.container.commitSHA"],
		}
		response = append(response, r)
	}
	return response, nil
}
