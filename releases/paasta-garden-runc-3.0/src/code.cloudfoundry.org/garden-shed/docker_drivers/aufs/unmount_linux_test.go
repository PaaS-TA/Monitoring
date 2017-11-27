package aufs_test

import (
	"io/ioutil"
	"os"
	"os/exec"

	"code.cloudfoundry.org/garden-shed/docker_drivers/aufs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnmountLinux", func() {
	var rootPath string

	BeforeEach(func() {
		var err error

		rootPath, err = ioutil.TempDir("", "mount-root")
		Expect(err).NotTo(HaveOccurred())

		Expect(exec.Command("mount", "-t", "tmpfs", "tmpfs", rootPath).Run()).To(Succeed())
	})

	Context("when unmount fails", func() {
		var file *os.File

		BeforeEach(func() {
			var err error

			file, err = ioutil.TempFile(rootPath, "data-file")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(file.Close()).To(Succeed())
			Expect(exec.Command("umount", rootPath).Run()).To(Succeed())
		})

		It("should return the error", func() {
			Expect(aufs.Unmount(rootPath)).To(HaveOccurred())
		})
	})

	Context("when mount succeeds", func() {
		Context("when path is not mount point", func() {
			BeforeEach(func() {
				Expect(exec.Command("umount", rootPath).Run()).To(Succeed())
			})

			It("should not return an error", func() {
				Expect(aufs.Unmount(rootPath)).NotTo(HaveOccurred())
			})
		})

		Context("when path is mount point", func() {
			It("should succeed to unmount", func() {
				Expect(aufs.Unmount(rootPath)).To(Succeed())
			})

			It("should not have an mount point", func() {
				Expect(aufs.Unmount(rootPath)).To(Succeed())
				Expect(exec.Command("mountpoint", rootPath).Run()).NotTo(Succeed())
			})
		})
	})
})
