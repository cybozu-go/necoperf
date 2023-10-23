package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

func podCandidates(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := k8sConfig.GetConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	pods := &corev1.PodList{}
	err = k8sClient.List(context.Background(), pods, client.InNamespace(config.namespace))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var podNames []string
	for _, pod := range pods.Items {
		if !strings.HasPrefix(pod.Name, toComplete) {
			continue
		}
		podNames = append(podNames, pod.Name)
	}

	return podNames, cobra.ShellCompDirectiveNoFileComp
}

func containerCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := k8sConfig.GetConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	pod := &corev1.Pod{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: config.namespace,
		Name:      args[0],
	}, pod)
	if err != nil {
		fmt.Println(err)
		return nil, cobra.ShellCompDirectiveError
	}

	var containerNames []string
	for _, c := range pod.Spec.Containers {
		if !strings.HasPrefix(c.Name, toComplete) {
			continue
		}
		containerNames = append(containerNames, c.Name)
	}

	return containerNames, cobra.ShellCompDirectiveNoFileComp
}

func namespaceCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := k8sConfig.GetConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	ns := &corev1.NamespaceList{}
	err = k8sClient.List(context.Background(), ns)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var namespaces []string
	for _, n := range ns.Items {
		if !strings.HasPrefix(n.Name, toComplete) {
			continue
		}
		namespaces = append(namespaces, n.Name)
	}

	return namespaces, cobra.ShellCompDirectiveNoFileComp
}
