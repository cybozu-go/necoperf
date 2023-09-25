package resource

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	apitesting "k8s.io/cri-api/pkg/apis/testing"
)

func TestGetPidFromContainerIDInvalidContainerID(t *testing.T) {
	t.Parallel()

	runningContainerID := "container-id"
	nonExistentContainerID := "non-existent-container-id"
	stoppingContainerID := "stopping-container-id"

	containers := map[string]*apitesting.FakeContainer{
		runningContainerID: {
			ContainerStatus: runtimeapi.ContainerStatus{
				State: runtimeapi.ContainerState_CONTAINER_RUNNING,
			},
		},
		stoppingContainerID: {
			ContainerStatus: runtimeapi.ContainerStatus{
				State: runtimeapi.ContainerState_CONTAINER_EXITED,
			},
		},
	}

	fakeRuntimeService := &apitesting.FakeRuntimeService{
		Containers: containers,
	}
	c := NewContainer(nil, fakeRuntimeService)

	_, err := c.GetPidFromContainerID(context.Background(), nonExistentContainerID)
	if err != nil {
		expected := fmt.Sprintf("container %q not found", nonExistentContainerID)
		assert.Equal(t, expected, err.Error())
	}

	_, err = c.GetPidFromContainerID(context.Background(), stoppingContainerID)
	if err != nil {
		expected := fmt.Sprintf("%q container is not running", stoppingContainerID)
		assert.Equal(t, expected, err.Error())
	}

	pid, err := c.GetPidFromContainerID(context.Background(), runningContainerID)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, pid)
}
