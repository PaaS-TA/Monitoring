package garden_integration_tests_test

import (
	"fmt"

	"code.cloudfoundry.org/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Devices", func() {
	DescribeTable("Devices",
		func(device string, major, minor int) {
			buffer := gbytes.NewBuffer()
			process, err := container.Run(garden.ProcessSpec{
				Path: "ls",
				Args: []string{"-l", device},
			}, garden.ProcessIO{Stdout: buffer, Stderr: GinkgoWriter})
			Expect(err).ToNot(HaveOccurred())

			exitCode, err := process.Wait()
			Expect(err).NotTo(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(buffer).To(gbytes.Say(fmt.Sprintf(`%d,\s*%d`, major, minor)))
		},

		Entry("should have the TTY device", "/dev/tty", 5, 0),
		Entry("should have the random device", "/dev/random", 1, 8),
		Entry("should have the urandom device", "/dev/urandom", 1, 9),
		Entry("should have the null device", "/dev/null", 1, 3),
		Entry("should have the zero device", "/dev/zero", 1, 5),
		Entry("should have the full device", "/dev/full", 1, 7),
		Entry("should have the /dev/pts/ptmx device", "/dev/pts/ptmx", 5, 2),
		Entry("should have the fuse device", "/dev/fuse", 10, 229),
	)

	DescribeTable("Process",
		func(device, fd string) {
			buffer := gbytes.NewBuffer()
			process, err := container.Run(garden.ProcessSpec{
				Path: "ls",
				Args: []string{"-l", device},
			}, garden.ProcessIO{Stdout: buffer, Stderr: GinkgoWriter})
			Expect(err).ToNot(HaveOccurred())

			exitCode, err := process.Wait()
			Expect(err).NotTo(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(buffer).To(gbytes.Say(fmt.Sprintf("%s -> %s", device, fd)))
		},
		Entry("should have /dev/fd", "/dev/fd", "/proc/self/fd"),
		Entry("should have /dev/stdin", "/dev/stdin", "/proc/self/fd/0"),
		Entry("should have /dev/stdout", "/dev/stdout", "/proc/self/fd/1"),
		Entry("should have /dev/stderr", "/dev/stderr", "/proc/self/fd/2"),
	)

	It("should have devpts mounted", func() {
		stdout := gbytes.NewBuffer()

		process, err := container.Run(garden.ProcessSpec{
			User: "root",
			Path: "cat",
			Args: []string{"/proc/mounts"},
		}, garden.ProcessIO{
			Stdout: stdout,
			Stderr: GinkgoWriter,
		})
		Expect(err).ToNot(HaveOccurred())

		exitCode, err := process.Wait()
		Expect(err).NotTo(HaveOccurred())
		Expect(exitCode).To(Equal(0))

		Expect(stdout).To(gbytes.Say("devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=666"))
	})
})
