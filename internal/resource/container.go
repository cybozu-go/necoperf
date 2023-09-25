package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	criapi "k8s.io/cri-api/pkg/apis"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
)

type Container struct {
	logger    *slog.Logger
	criClient criapi.RuntimeService
}

type containerStatus struct {
	PID int `json:"pid"`
}

func NewContainer(logger *slog.Logger, criClient criapi.RuntimeService) *Container {
	return &Container{
		criClient: criClient,
		logger:    logger,
	}
}

// GetPidFromContainerID returns the pid of the container
func (c *Container) GetPidFromContainerID(ctx context.Context, containerID string) (int, error) {
	var status containerStatus

	resp, err := c.criClient.ContainerStatus(ctx, containerID, true)
	if err != nil {
		return -1, err
	}

	if resp.Status.State != runtimeapi.ContainerState_CONTAINER_RUNNING {
		return -1, fmt.Errorf("%q container is not running", containerID)
	}

	for k := range resp.Info {
		if err := json.Unmarshal([]byte(resp.Info[k]), &status); err != nil {
			return -1, err
		}
	}

	return status.PID, nil
}
