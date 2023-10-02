package resource

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/cybozu-go/necoperf/internal/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var cancelCluster context.CancelFunc
var testEnv *envtest.Environment
var scheme = runtime.NewScheme()
var k8sClient client.Client
var testThreshold time.Duration
var d *Discovery

const (
	HostIP = "10.69.0.197"
)

func TestDiscovery(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(1 * time.Minute)
	SetDefaultEventuallyPollingInterval(1 * time.Second)
	RunSpecs(t, "Test discovery")
}

var _ = BeforeSuite(func() {
	var err error

	testThreshold, err = time.ParseDuration("5s")
	Expect(err).NotTo(HaveOccurred())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{}

	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	var ctx context.Context
	ctx, cancelCluster = context.WithCancel(context.Background())
	testNamespace := corev1.Namespace{}
	testNamespace.Name = "test"

	testPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "t1",
					Image: "necoperf",
				},
			},
		},
	}

	necoperfNamespace := corev1.Namespace{}
	necoperfNamespace.Name = "test-for-necoperf"
	testDaemonset := appsv1.DaemonSet{}
	testDaemonset.Namespace = "test-for-necoperf"
	testDaemonset.ObjectMeta.Name = "necoperf-daemon"
	testDaemonset.ObjectMeta.Labels = map[string]string{constants.LabelAppName: constants.AppNameNecoPerf}
	testDaemonset.Spec.Selector = &metav1.LabelSelector{MatchLabels: map[string]string{constants.LabelAppName: constants.AppNameNecoPerf}}
	testDaemonset.Spec.Template.ObjectMeta.Labels = map[string]string{constants.LabelAppName: constants.AppNameNecoPerf}
	testDaemonset.Spec.Template.Spec.Containers = []corev1.Container{{Name: "n1", Image: "necoperf"}}

	err = k8sClient.Create(ctx, &testNamespace)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Create(ctx, &testPod)
	Expect(err).NotTo(HaveOccurred())
	updatePodStatus(ctx, &testPod, "containerd://necoperf", "t1", "10.244.1.2",
		corev1.ContainerState{Running: &corev1.ContainerStateRunning{}})

	err = k8sClient.Create(ctx, &necoperfNamespace)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Create(ctx, &testDaemonset)
	Expect(err).NotTo(HaveOccurred())
	updateDaemonSetStatus(ctx, &testDaemonset)

	d, err = NewDiscovery(slog.Default(), k8sClient)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	cancelCluster()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func updatePodStatus(ctx context.Context, pod *corev1.Pod, containerID, name, podIP string, state corev1.ContainerState) {
	pod.Status.ContainerStatuses = []corev1.ContainerStatus{
		{
			ContainerID: containerID,
			Name:        name,
			State:       state,
		},
	}
	pod.Status.HostIP = HostIP
	pod.Status.PodIP = podIP
	err := k8sClient.Status().Update(ctx, pod)
	Expect(err).NotTo(HaveOccurred())
}

func updateDaemonSetStatus(ctx context.Context, ds *appsv1.DaemonSet) {
	ds.Status.DesiredNumberScheduled = 1
	ds.Status.NumberAvailable = 1
	ds.Status.NumberReady = 1
	ds.Status.NumberUnavailable = 0
	err := k8sClient.Status().Update(ctx, ds)
	Expect(err).NotTo(HaveOccurred())

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: ds.GetObjectMeta().GetName() + "1", Namespace: ds.GetNamespace(), Labels: ds.Spec.Template.GetObjectMeta().GetLabels()},
		Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: "n1", Image: "necoperf"},
		},
		},
	}
	pod.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(ds, appsv1.SchemeGroupVersion.WithKind("DaemonSet"))}
	err = k8sClient.Create(ctx, pod)
	Expect(err).NotTo(HaveOccurred())

	updatePodStatus(ctx, pod, "containerd://necoperf", "necoperf-daemon", "10.244.1.3",
		corev1.ContainerState{Running: &corev1.ContainerStateRunning{}})
}
