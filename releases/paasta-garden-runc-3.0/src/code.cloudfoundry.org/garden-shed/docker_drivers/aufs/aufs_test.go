package aufs_test

import (
	"errors"

	"code.cloudfoundry.org/garden-shed/docker_drivers/aufs"
	fakes "code.cloudfoundry.org/garden-shed/docker_drivers/aufs/aufsfakes"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuotaedDriver", func() {
	var (
		fakeGraphDriver     *fakes.FakeGraphDriver
		fakeLoopMounter     *fakes.FakeLoopMounter
		fakeBackingStoreMgr *fakes.FakeBackingStoreMgr
		fakeRetrier         *fakes.FakeRetrier
		fakeUnmount         aufs.UnmountFunc

		driver *aufs.QuotaedDriver

		rootPath string
	)

	BeforeEach(func() {
		rootPath = "/path/to/my/banana/graph"

		fakeGraphDriver = new(fakes.FakeGraphDriver)
		fakeLoopMounter = new(fakes.FakeLoopMounter)
		fakeBackingStoreMgr = new(fakes.FakeBackingStoreMgr)
		fakeRetrier = new(fakes.FakeRetrier)
		fakeUnmount = func(path string) error {
			return nil
		}
	})

	JustBeforeEach(func() {
		driver = &aufs.QuotaedDriver{
			GraphDriver:     fakeGraphDriver,
			Unmount:         fakeUnmount,
			BackingStoreMgr: fakeBackingStoreMgr,
			LoopMounter:     fakeLoopMounter,
			Retrier:         fakeRetrier,
			RootPath:        rootPath,
			Logger:          lagertest.NewTestLogger("test"),
		}
	})

	Describe("GetQuotaed", func() {
		It("should create a backing store file", func() {
			id := "banana-id"
			quota := int64(12 * 1024)

			_, err := driver.GetQuotaed(id, "", quota)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeBackingStoreMgr.CreateCallCount()).To(Equal(1))
			gottenId, gottenQuota := fakeBackingStoreMgr.CreateArgsForCall(0)
			Expect(gottenId).To(Equal(id))
			Expect(gottenQuota).To(Equal(quota))
		})

		Context("when failing to create a backing store", func() {
			It("should return an error", func() {
				fakeBackingStoreMgr.CreateReturns("", errors.New("create failed!"))

				_, err := driver.GetQuotaed("banana-id", "", 12*1024)
				Expect(err).To(MatchError(ContainSubstring("create failed!")))
			})
		})

		It("should mount the backing store file", func() {
			realDevicePath := "/path/to/my/banana/device"

			fakeBackingStoreMgr.CreateReturns(realDevicePath, nil)

			_, err := driver.GetQuotaed("banana-id", "", 10*1024)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLoopMounter.MountFileCallCount()).To(Equal(1))
			devicePath, destPath := fakeLoopMounter.MountFileArgsForCall(0)
			Expect(devicePath).To(Equal(realDevicePath))
			Expect(destPath).To(Equal("/path/to/my/banana/graph/aufs/diff/banana-id"))
		})

		Context("when failing to mount the backing store", func() {
			BeforeEach(func() {
				fakeLoopMounter.MountFileReturns(errors.New("another banana error"))
			})

			It("should return an error", func() {
				_, err := driver.GetQuotaed("banana-id", "", 10*1024)
				Expect(err).To(MatchError(ContainSubstring("another banana error")))
			})

			It("should not mount the layer", func() {
				driver.GetQuotaed("banana-id", "", 10*1024*1024)
				Expect(fakeGraphDriver.GetCallCount()).To(Equal(0))
			})
		})

		It("should mount the layer", func() {
			id := "mango-id"
			mountLabel := "wild mangos: handle with care"
			quota := int64(12 * 1024 * 1024)

			_, err := driver.GetQuotaed(id, mountLabel, quota)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeGraphDriver.GetCallCount()).To(Equal(1))
			gottenID, gottenMountLabel := fakeGraphDriver.GetArgsForCall(0)
			Expect(gottenID).To(Equal(id))
			Expect(gottenMountLabel).To(Equal(mountLabel))
		})

		It("should return the mounted layer's path", func() {
			mountPath := "/path/to/mounted/banana"

			fakeGraphDriver.GetReturns(mountPath, nil)

			path, err := driver.GetQuotaed("test-banana-id", "", 10*1024*1024)
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal(mountPath))
		})

		Context("when mounting the layer fails", func() {
			It("should return an error", func() {
				fakeGraphDriver.GetReturns("", errors.New("Another banana error"))

				_, err := driver.GetQuotaed("banana-id", "", 10*1024*1024)
				Expect(err).To(MatchError(ContainSubstring("Another banana error")))
			})
		})
	})

	Describe("Put", func() {
		It("should put the layer", func() {
			id := "herring-id"

			Expect(driver.Put(id)).To(Succeed())

			Expect(fakeGraphDriver.PutCallCount()).To(Equal(1))
			Expect(fakeGraphDriver.PutArgsForCall(0)).To(Equal(id))
		})

		It("should retry unmounting the mnt endpoint", func() {})

		Context("when the retrier fails", func() {
			BeforeEach(func() {
				fakeRetrier.RunReturns(errors.New("banana"))
			})

			It("should return the error", func() {
				Expect(driver.Put("banana-magic")).To(MatchError("banana"))
			})

			It("should not unmount the loop device", func() {
				Expect(driver.Put("banana-magic")).To(HaveOccurred())
				Expect(fakeLoopMounter.UnmountCallCount()).To(Equal(0))
			})
		})

		It("should unmount the loop mount", func() {
			Expect(driver.Put("banana-id")).To(Succeed())

			Expect(fakeLoopMounter.UnmountCallCount()).To(Equal(1))
			Expect(fakeLoopMounter.UnmountArgsForCall(0)).To(Equal("/path/to/my/banana/graph/aufs/diff/banana-id"))
		})

		Context("when unmounting the loop device fails", func() {
			BeforeEach(func() {
				fakeLoopMounter.UnmountReturns(errors.New("avocado"))
			})

			It("should return an error", func() {
				Expect(driver.Put("banana-id")).To(MatchError("unmounting the loop device: avocado"))
			})

			It("should not remove the backing store file", func() {
				Expect(driver.Put("banana-id")).To(HaveOccurred())
				Expect(fakeBackingStoreMgr.DeleteCallCount()).To(Equal(0))
			})
		})

		It("should remove the backing store file", func() {
			id := "banana-id"

			driver.GetQuotaed(id, "", 10*1024)

			Expect(driver.Put("banana-id")).To(Succeed())

			Expect(fakeRetrier.RunCallCount()).To(Equal(1))

			Expect(fakeBackingStoreMgr.DeleteCallCount()).To(Equal(1))
			Expect(fakeBackingStoreMgr.DeleteArgsForCall(0)).To(Equal(id))
		})

		Context("when removig the backing store file fails", func() {
			BeforeEach(func() {
				fakeBackingStoreMgr.DeleteReturns(errors.New("banana"))
			})

			It("should return an error", func() {
				Expect(driver.Put("banana-shed")).To(MatchError("removing the backing store: banana"))
			})
		})
	})

	Describe("GetMntPath", func() {
		It("returns the mnt path of the given layer (without calling Path)", func() {
			Expect(driver.GetMntPath(layercake.DockerImageID("foo"))).To(Equal("/path/to/my/banana/graph/aufs/mnt/foo"))
		})
	})

	Describe("GetDiffLayerPath", func() {
		It("replaces the `/mnt/ substring to `/diff/ in the given path", func() {
			path := "some/mnt/path"
			Expect(driver.GetDiffLayerPath(path)).To(Equal("some/diff/path"))
		})

		It("only replaces the first occurance of `/mnt/` to `/diff/` in the given path", func() {
			path := "some/mnt/mnt/path"
			Expect(driver.GetDiffLayerPath(path)).To(Equal("some/diff/mnt/path"))
		})
	})
})
