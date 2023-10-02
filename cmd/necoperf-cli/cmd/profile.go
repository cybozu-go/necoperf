package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cybozu-go/necoperf/internal/client"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var config struct {
	namespace     string
	podName       string
	outputDir     string
	containerName string
	necoperfNS    string
	timeout       time.Duration
}

func isPodReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type != corev1.PodReady {
			continue
		}
		return cond.Status == corev1.ConditionTrue
	}
	return false
}

func NewProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Exec perf profile of the target container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config.podName = args[0]
			handler := slog.NewTextHandler(os.Stderr, nil)
			logger := slog.New(handler)

			client, err := client.New(logger, config.timeout)
			if err != nil {
				return err
			}
			if err := client.SetupDiscovery(); err != nil {
				return err
			}

			ctx := cmd.Context()
			pod, err := client.Discovery.GetPod(ctx, config.namespace, config.podName)
			if err != nil {
				return err
			}
			if !isPodReady(pod) {
				return fmt.Errorf("pod %s is not ready", pod.Name)
			}
			containerID, err := client.Discovery.GetContainerID(pod, config.containerName)
			if err != nil {
				return err
			}
			logger.Info("get container id", "podName", config.podName, "containerID", containerID)

			pods, err := client.Discovery.GetPodList(ctx, config.necoperfNS)
			if err != nil {
				return err
			}
			addr, err := client.Discovery.DiscoveryServerAddr(pods, pod.Status.HostIP)
			if err != nil {
				return err
			}
			err = client.SetupGrpcClient(addr)
			if err != nil {
				return err
			}
			logger.Info("connect grpc server", "addr", addr)

			err = client.Profile(ctx, config.podName, containerID, config.outputDir)
			if err != nil {
				return err
			}
			logger.Info("profile is finished", "output directory", filepath.Join(config.outputDir, config.podName+".script"))

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "kubernetes namespace")
	cmd.Flags().StringVar(&config.outputDir, "outputDir", "/tmp", "output data directory")
	cmd.Flags().StringVar(&config.containerName, "container", "", "container name")
	cmd.Flags().StringVar(&config.necoperfNS, "necoperf-namespace", "kube-system", "necoperf namespace")
	cmd.Flags().DurationVar(&config.timeout, "timeout", time.Second*30, "timeout seconds")

	return cmd
}
