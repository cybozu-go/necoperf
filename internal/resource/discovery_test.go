package resource

import (
	"context"
	"fmt"

	"github.com/cybozu-go/necoperf/internal/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Discovery", func() {
	ctx := context.Background()

	It("should get container id", func() {
		pod, err := d.GetPod(ctx, "test", "test-pod")
		Expect(err).NotTo(HaveOccurred())
		Expect(pod.Status.ContainerStatuses[0].ContainerID).NotTo(BeEmpty())
		containerID, err := d.GetContainerID(pod, "t1")
		Expect(err).NotTo(HaveOccurred())
		Expect(containerID).NotTo(BeEmpty())
	})

	It("should discovery server addr", func() {
		By("get test pod")
		pod, err := d.GetPod(ctx, "test", "test-pod")
		Expect(err).NotTo(HaveOccurred())
		Expect(pod.Status.HostIP).To(Equal(HostIP))

		By("get test daemonset")
		podList, err := d.GetPodList(ctx, "test-default-port")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))

		By("discovery server addr")
		addr, err := d.DiscoveryServerAddr(podList, pod.Status.HostIP)
		Expect(err).NotTo(HaveOccurred())
		Expect(addr).To(Equal(fmt.Sprintf("%s:%d", daemonsetPodIP, constants.NecoPerfGrpcServerPort)))
	})

	It("should get the port specified by the server when the server specifies a port other than the default port", func() {
		By("get test pod")
		pod, err := d.GetPod(ctx, "test", "test-pod")
		Expect(err).NotTo(HaveOccurred())
		Expect(pod.Status.HostIP).To(Equal(HostIP))

		By("get specified port daemonset")
		podList, err := d.GetPodList(ctx, "test-specified-port")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(podList.Items)).To(Equal(1))

		By("discovery server addr")
		addr, err := d.DiscoveryServerAddr(podList, pod.Status.HostIP)
		Expect(err).NotTo(HaveOccurred())
		Expect(addr).To(Equal(fmt.Sprintf("%s:%d", daemonsetPodIP, 8080)))
	})
})
