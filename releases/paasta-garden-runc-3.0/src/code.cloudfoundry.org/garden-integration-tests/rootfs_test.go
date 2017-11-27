package garden_integration_tests_test

import (
	"io"
	"os"

	"code.cloudfoundry.org/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Rootfses", func() {
	BeforeEach(func() {
		imageRef.URI = "docker:///cfgarden/with-volume"
	})

	Context("when the rootfs path is a docker image URL", func() {
		Context("and the image specifies $PATH", func() {
			It("$PATH is taken from the docker image", func() {
				stdout := gbytes.NewBuffer()
				process, err := container.Run(garden.ProcessSpec{
					User: "root",
					Path: "/bin/sh",
					Args: []string{"-c", "echo $PATH"},
				}, garden.ProcessIO{
					Stdout: io.MultiWriter(GinkgoWriter, stdout),
					Stderr: GinkgoWriter,
				})

				Expect(err).ToNot(HaveOccurred())

				process.Wait()
				Expect(stdout).To(gbytes.Say("/usr/local/bin:/usr/bin:/bin:/from-dockerfile"))
			})

			It("$TEST is taken from the docker image", func() {
				stdout := gbytes.NewBuffer()
				process, err := container.Run(garden.ProcessSpec{
					User: "root",
					Path: "/bin/sh",
					Args: []string{"-c", "echo $TEST"},
				}, garden.ProcessIO{
					Stdout: io.MultiWriter(GinkgoWriter, stdout),
					Stderr: GinkgoWriter,
				})

				Expect(err).ToNot(HaveOccurred())

				process.Wait()
				Expect(stdout).To(gbytes.Say("second-test-from-dockerfile:test-from-dockerfile"))
			})

		})

		Context("and the image specifies a VOLUME", func() {
			It("creates the volume directory, if it does not already exist", func() {
				stdout := gbytes.NewBuffer()
				process, err := container.Run(garden.ProcessSpec{
					User: "root",
					Path: "ls",
					Args: []string{"-l", "/"},
				}, garden.ProcessIO{
					Stdout: io.MultiWriter(GinkgoWriter, stdout),
					Stderr: GinkgoWriter,
				})

				Expect(err).ToNot(HaveOccurred())
				process.Wait()
				Expect(stdout).To(gbytes.Say("foo"))
			})
		})

		Context("and the image is private", func() {
			BeforeEach(func() {
				imageRef.URI = "docker:///cfgarden/private"
				imageRef.Username = os.Getenv("REGISTRY_USERNAME")
				imageRef.Password = os.Getenv("REGISTRY_PASSWORD")
				if imageRef.Username == "" || imageRef.Password == "" {
					Skip("Registry username or password not provided")
				}
				assertContainerCreate = false
			})

			It("successfully pulls the image", func() {
				Expect(containerCreateErr).ToNot(HaveOccurred())
			})

			Context("but the credentials are incorrect", func() {
				BeforeEach(func() {
					imageRef.Username = ""
					imageRef.Password = ""
				})

				It("fails", func() {
					Expect(containerCreateErr).NotTo(Succeed())
				})
			})
		})
	})
})
