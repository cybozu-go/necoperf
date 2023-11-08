package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("necoperf e2e test", func() {
	It("should be able to run necoperf-cli", func() {
		By("executing necoperf-cli")
		_, err := kubectl(nil, "exec", "necoperf-client", "--",
			"necoperf-cli", "profile", "profiled-pod",
			"--timeout", "100s")
		Expect(err).NotTo(HaveOccurred())

		By("checking if profiling result is created")
		out, err := kubectl(nil, "exec", "necoperf-client", "--", "cat", "/tmp/profiled-pod.script")
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("yes"))
	})
})
