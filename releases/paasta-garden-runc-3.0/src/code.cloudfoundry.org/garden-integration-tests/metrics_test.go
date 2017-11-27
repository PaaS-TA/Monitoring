package garden_integration_tests_test

import (
	"code.cloudfoundry.org/garden"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics", func() {
	JustBeforeEach(func() {
		_, err := container.Run(garden.ProcessSpec{
			Path: "sh",
			Args: []string{
				"-c", `while true; do; ls -la; done`,
			},
		}, garden.ProcessIO{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should return the CPU metrics", func() {
		Eventually(func() uint64 {
			metrics, err := container.Metrics()
			Expect(err).NotTo(HaveOccurred())

			return metrics.CPUStat.Usage
		}).ShouldNot(BeZero())
	})

	It("should return the memory metrics", func() {
		Eventually(func() uint64 {
			metrics, err := container.Metrics()
			Expect(err).NotTo(HaveOccurred())

			return metrics.MemoryStat.TotalUsageTowardLimit
		}).ShouldNot(BeZero())
	})

	It("should return bulk metrics", func() {
		metrics, err := gardenClient.BulkMetrics([]string{container.Handle()})
		Expect(err).NotTo(HaveOccurred())

		Expect(metrics).To(HaveKey(container.Handle()))
		Expect(metrics[container.Handle()].Metrics.MemoryStat.TotalUsageTowardLimit).NotTo(BeZero())
	})
})
