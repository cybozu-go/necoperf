package cmd

import (
	"context"
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
		Use:               "profile",
		Short:             "Perform CPU profiling on the target container",
		Long:              "Perform CPU profiling on the target container",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: podCandidates,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			config.podName = args[0]
			handler := slog.NewTextHandler(os.Stderr, nil)
			logger := slog.New(handler)

			client, err := client.New(logger, config.timeout)
			if err != nil {
				return err
			}
			ds, err := client.SetupDiscovery()
			if err != nil {
				return err
			}

			ctx := context.Background()
			pod, err := ds.GetPod(ctx, config.namespace, config.podName)
			if err != nil {
				return err
			}
			if !isPodReady(pod) {
				return fmt.Errorf("pod %s is not ready", pod.Name)
			}
			containerID, err := ds.GetContainerID(pod, config.containerName)
			if err != nil {
				return err
			}
			logger.Info("get container id", "podName", config.podName, "containerID", containerID)

			pods, err := ds.GetPodList(ctx, config.necoperfNS)
			if err != nil {
				return err
			}
			addr, err := ds.DiscoveryServerAddr(pods, pod.Status.HostIP)
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
	}
	cmd.Flags().StringVar(&config.necoperfNS, "necoperf-namespace", "necoperf", "Namespace in which necoperf-daemon is running")
	cmd.Flags().StringVarP(&config.namespace, "namespace", "n", "default", "Namespace in pod being profiled is running")
	cmd.Flags().StringVarP(&config.containerName, "container", "c", "", "Specify the container name to profile")
	cmd.Flags().DurationVar(&config.timeout, "timeout", 30*time.Second, "Time to run cpu profiling on server")
	cmd.Flags().StringVar(&config.outputDir, "output-dir", "/tmp", "Directory to output profiling result")
	cmd.RegisterFlagCompletionFunc("namespace", namespaceCompletionFunc)
	cmd.RegisterFlagCompletionFunc("container", containerCompletionFunc)

	return cmd
}
