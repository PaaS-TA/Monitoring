package garden_integration_tests_test

import (
	"bytes"
	"io"
	"strconv"

	"code.cloudfoundry.org/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("users", func() {
	Context("when nobody maps to 65534", func() {
		BeforeEach(func() {
			imageRef.URI = "docker:///ubuntu"
		})

		It("should be able to su to nobody", func() {
			// Delete guff.
			proc, err := container.Run(garden.ProcessSpec{
				User: "root",
				Path: "rm",
				Args: []string{"-f", "/etc/passwd-"},
			}, garden.ProcessIO{
				Stdout: GinkgoWriter,
				Stderr: GinkgoWriter,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))

			// Ensure "nobody" has a shell.
			proc, err = container.Run(garden.ProcessSpec{
				User: "root",
				Path: "chsh",
				Args: []string{"-s", "/bin/bash", "nobody"},
			}, garden.ProcessIO{
				Stdout: GinkgoWriter,
				Stderr: GinkgoWriter,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))

			var buf bytes.Buffer
			proc, err = container.Run(garden.ProcessSpec{
				User: "root",
				Path: "su",
				Args: []string{"nobody", "-c", "whoami"},
			}, garden.ProcessIO{
				Stdout: io.MultiWriter(&buf, GinkgoWriter),
				Stderr: io.MultiWriter(&buf, GinkgoWriter),
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))
			Expect(buf.String()).To(Equal("nobody\n"))
		})
	})

	Context("when creating users", func() {
		BeforeEach(func() {
			imageRef.URI = "docker:///cfgarden/garden-busybox"
		})

		It("creates a user with a large uid and gid", func() {
			uid := 700000
			gid := 700000

			proc, err := container.Run(garden.ProcessSpec{
				User: "root",
				Path: "addgroup",
				Args: []string{"-g", strconv.Itoa(gid), "bob"},
			}, garden.ProcessIO{
				Stdout: GinkgoWriter,
				Stderr: GinkgoWriter,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))

			proc, err = container.Run(garden.ProcessSpec{
				User: "root",
				Path: "adduser",
				Args: []string{"-u", strconv.Itoa(uid), "-G", "bob", "-D", "bob"},
			}, garden.ProcessIO{
				Stdout: GinkgoWriter,
				Stderr: GinkgoWriter,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))

			proc, err = container.Run(garden.ProcessSpec{
				User: "bob",
				Path: "echo",
				Args: []string{"Hello Baldrick"},
			}, garden.ProcessIO{
				Stdout: GinkgoWriter,
				Stderr: GinkgoWriter,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))
		})
	})

	Context("when rootfs defines user/groups", func() {
		BeforeEach(func() {
			imageRef.URI = "docker:///cfgarden/with-user-with-groups"
		})

		It("ignores additional groups", func() {
			stdout := gbytes.NewBuffer()

			proc, err := container.Run(garden.ProcessSpec{
				User: "alice",
				Path: "cat",
				Args: []string{"/proc/self/status"},
			}, garden.ProcessIO{
				Stdout: stdout,
				Stderr: GinkgoWriter,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(proc.Wait()).To(Equal(0))
			Expect(stdout).To(gbytes.Say("Groups:\t\n"))
			Expect(stdout).NotTo(gbytes.Say("1010"))
			Expect(stdout).NotTo(gbytes.Say("1011"))
		})
	})
})
