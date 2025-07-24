package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
)

// Version returns the Docker server version
func (c *Client) Version(ctx context.Context) (*types.Version, error) {
	version, err := c.ServerVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker server version: %w", err)
	}
	return &version, nil
}
