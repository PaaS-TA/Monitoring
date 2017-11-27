package garden_integration_tests_test

import (
	"code.cloudfoundry.org/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("smoke tests", func() {
	It("can run a process inside a container", func() {
		stdout := gbytes.NewBuffer()

		_, err := container.Run(garden.ProcessSpec{
			Path: "whoami",
			User: "root",
		}, garden.ProcessIO{
			Stdout: stdout,
			Stderr: GinkgoWriter,
		})

		Expect(err).ToNot(HaveOccurred())
		Eventually(stdout).Should(gbytes.Say("root\n"))
	})
})
