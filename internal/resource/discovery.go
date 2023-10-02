package resource

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"

	"github.com/cybozu-go/necoperf/internal/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	port = 6543
)

type Discovery struct {
	logger *slog.Logger
	client client.Client
}

func NewDiscovery(logger *slog.Logger, client client.Client) (*Discovery, error) {
	return &Discovery{
		logger: logger,
		client: client,
	}, nil
}

func (d *Discovery) GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	err := d.client.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      podName,
	}, pod)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (d *Discovery) GetPodList(ctx context.Context, necoperfNS string) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	err := d.client.List(ctx, pods, client.InNamespace(necoperfNS), client.MatchingLabels{
		constants.LabelAppName: constants.AppNameNecoPerf,
	})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func (d *Discovery) GetContainerID(pod *corev1.Pod, containerName string) (string, error) {
	if len(containerName) == 0 && len(pod.Status.Conditions) >= 1 {
		containerName = pod.Spec.Containers[0].Name
	}

	for i := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[i].Name == containerName {
			regex := regexp.MustCompile("[a-z]*://")
			containerID := regex.ReplaceAllString(pod.Status.ContainerStatuses[i].ContainerID, "")
			return containerID, nil
		}
	}

	return "", errors.New("failed to get container ID")
}

func (d *Discovery) DiscoveryServerAddr(pods *corev1.PodList, hostIP string) (string, error) {
	var podIP string
	for _, pod := range pods.Items {
		if pod.Status.HostIP == hostIP {
			podIP = pod.Status.PodIP
		}
	}
	if len(podIP) == 0 {
		return "", errors.New("failed to get pod IP")
	}

	addr := fmt.Sprintf("%s:%s", podIP, strconv.Itoa(port))
	d.logger.Info("found the pod IP of necoperf", "hostIP", hostIP, "podIP", podIP)

	return addr, nil
}
