package e2e

import (
	"encoding/json"
	"fmt"
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
	Eventually(func() error {
		res, err := kubectl(nil, "get", "pod", "profiled-pod", "-o", "json")
		if err != nil {
			return err
		}

		pod := corev1.Pod{}
		err = json.Unmarshal(res, &pod)
		if err != nil {
			return err
		}
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("profiled-pod is not running")
		}

		return nil
	}).Should(Succeed())

	By("creating necoperf-cli pod from manifests")
	_, err = kubectl(nil, "apply", "-f", "./manifests/necoperf-client.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("waiting for necoperf-cli pod to be ready")
	Eventually(func() error {
		res, err := kubectl(nil, "get", "pod", "necoperf-client", "-o", "json")
		if err != nil {
			return err
		}

		pod := corev1.Pod{}
		err = json.Unmarshal(res, &pod)
		if err != nil {
			return err
		}
		if pod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("necoperf-client pod is not running")
		}

		return nil
	}).Should(Succeed())

	By("creating necoperf daemonset from manifest")
	_, err = kubectl(nil, "apply", "-f", "./manifests/necoperf-daemonset.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("waiting for necoperf-daemon daemonnset to be ready")
	Eventually(func() error {
		res, err := kubectl(nil, "get", "daemonset", "necoperf-daemon", "-n", "necoperf", "-o", "json")
		if err != nil {
			return err
		}

		daemonset := appsv1.DaemonSet{}
		err = json.Unmarshal(res, &daemonset)
		if err != nil {
			return err
		}

		if int(daemonset.Status.NumberAvailable) != daemonsetDesiredNumber {
			return fmt.Errorf("NumberAvailable of necoperf DaemonSet is not %d: %d", daemonsetDesiredNumber, daemonset.Status.NumberAvailable)
		}
		if int(daemonset.Status.UpdatedNumberScheduled) != daemonsetDesiredNumber {
			return fmt.Errorf("UpdatedNumberScheduled of necoperf DaemonSet is not %d: %d", daemonsetDesiredNumber, daemonset.Status.UpdatedNumberScheduled)
		}
		if int(daemonset.Status.NumberReady) != daemonsetDesiredNumber {
			return fmt.Errorf("NumberReady of necoperf DaemonSet is not %d: %d", daemonsetDesiredNumber, daemonset.Status.NumberReady)
		}

		return nil
	}, 3*time.Minute).Should(Succeed())
})
