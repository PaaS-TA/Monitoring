package aufs_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/garden-shed/docker_drivers/aufs"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("BackingStoreLinux", func() {
	var (
		mgr      *aufs.BackingStore
		rootPath string
	)

	BeforeEach(func() {
		var err error

		rootPath, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		mgr = &aufs.BackingStore{
			RootPath: rootPath,
			Logger:   lagertest.NewTestLogger("test"),
		}
	})

	AfterEach(func() {
		Expect(os.RemoveAll(rootPath)).To(Succeed())
	})

	Describe("Create", func() {
		It("should return a path to an existing file in the provided root", func() {
			path, err := mgr.Create("banana_id", 1024*1024*20)
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Dir(path)).To(Equal(rootPath))
			Expect(path).To(BeAnExistingFile())
		})

		Context("when the RootPath does not exist", func() {
			BeforeEach(func() {
				rootPath = "/invalid/banana/rootpath"
			})

			It("should return a sensible error", func() {
				_, err := mgr.Create("banana-id", 1024*1024*20)
				Expect(err).To(MatchError(ContainSubstring("creating the backing store file")))
			})
		})

		It("should apply the provided quota", func() {
			quota := int64(10 * 1024 * 1024)
			path, err := mgr.Create("banana_id", quota)
			Expect(err).NotTo(HaveOccurred())

			fi, err := os.Stat(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.Size()).To(Equal(quota))
		})

		Context("when the quota is negative", func() {
			It("should return an error", func() {
				_, err := mgr.Create("banana-id", -12)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the quota is zero", func() {
			It("should return a sensible error message", func() {
				_, err := mgr.Create("awesome", 0)
				Expect(err).To(MatchError("cannot have zero sized quota"))
			})
		})

		It("should format the file as ext4", func() {
			path, err := mgr.Create("banana_id", 10*1024*1024)
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(
				exec.Command("blkid", path), GinkgoWriter, GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("TYPE=\"ext4\""))
		})

		It("should create an unjournaled backing store", func() {
			path, err := mgr.Create("banana_id", 10*1024*1024)
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(
				exec.Command("dumpe2fs", path), GinkgoWriter, GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(session).NotTo(gbytes.Say("has_journal"))
		})

		Context("when the quota is not enought", func() {
			It("should return an error", func() {
				quota := int64(1 * 1024)
				_, err := mgr.Create("banana_id", quota)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete", func() {
		It("should delete the file associated with the provided id", func() {
			id := "banana_id"
			path, err := mgr.Create(id, 1024*1024*20)
			Expect(err).NotTo(HaveOccurred())

			Expect(mgr.Delete(id)).To(Succeed())

			Expect(path).NotTo(BeAnExistingFile())
		})

		Context("when there is no file for the provided id", func() {
			It("should succeed", func() {
				Expect(mgr.Delete("fake-banana-id")).To(Succeed())
			})
		})
	})
})
