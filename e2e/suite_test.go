package e2e

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	daemonsetDesiredNumber = 1
)

func TestE2E(t *testing.T) {
	if !runE2E {
		t.Skip("no RUN_E2E environment variable")
	}

	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(1 * time.Minute)
	SetDefaultEventuallyPollingInterval(1 * time.Second)
	RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func() {
	By("prepare pod for profiling")
	_, err := kubectl(nil, "apply", "-f", "./manifests/profiled-pod.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("waiting for profiled-pod to be ready")
	Eventually(func(g Gomega) {
		res, err := kubectl(nil, "get", "pod", "profiled-pod", "-o", "json")
		g.Expect(err).NotTo(HaveOccurred())
		pod := corev1.Pod{}
		err = json.Unmarshal(res, &pod)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
	}).Should(Succeed())

	By("creating necoperf-cli pod from manifests")
	_, err = kubectl(nil, "apply", "-f", "./manifests/necoperf-client.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("waiting for necoperf-cli pod to be ready")
	Eventually(func(g Gomega) {
		res, err := kubectl(nil, "get", "pod", "necoperf-client", "-o", "json")
		g.Expect(err).NotTo(HaveOccurred())

		pod := corev1.Pod{}
		err = json.Unmarshal(res, &pod)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(pod.Status.Phase).To(Equal(corev1.PodRunning))
	}).Should(Succeed())

	By("creating necoperf daemonset from manifest")
	_, err = kubectl(nil, "apply", "-f", "./manifests/necoperf-daemonset.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("waiting for necoperf-daemon daemonnset to be ready")
	Eventually(func(g Gomega) {
		res, err := kubectl(nil, "get", "daemonset", "necoperf-daemon", "-n", "necoperf", "-o", "json")
		g.Expect(err).NotTo(HaveOccurred())
		daemonset := appsv1.DaemonSet{}
		err = json.Unmarshal(res, &daemonset)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(int(daemonset.Status.NumberAvailable)).To(Equal(daemonsetDesiredNumber))
		g.Expect(int(daemonset.Status.UpdatedNumberScheduled)).To(Equal(daemonsetDesiredNumber))
		g.Expect(int(daemonset.Status.NumberReady)).To(Equal(daemonsetDesiredNumber))
	}, 3*time.Minute).Should(Succeed())
})
