package aufs_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/garden-shed/docker_drivers/aufs"
	fakes "code.cloudfoundry.org/garden-shed/docker_drivers/aufs/aufsfakes"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const (
	pollingInterval = 100 * time.Millisecond
	numRetries      = 3 // At least 3 so that we can test the case where it works after multiple (2) retries
)

var _ = Describe("LoopLinux", func() {
	var (
		bsFilePath string
		destPath   string
		loop       *aufs.Loop

		fakeRetrier *fakes.FakeRetrier
	)

	BeforeEach(func() {
		var err error

		tempFile, err := ioutil.TempFile("", "loop")
		Expect(err).NotTo(HaveOccurred())
		bsFilePath = tempFile.Name()
		_, err = exec.Command("truncate", "-s", "10M", bsFilePath).CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		_, err = exec.Command("mkfs.ext4", "-F", bsFilePath).CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		destPath, err = ioutil.TempDir("", "loop")
		Expect(err).NotTo(HaveOccurred())

		fakeRetrier = new(fakes.FakeRetrier)
		fakeRetrier.RunStub = func(fn func() error) error {
			return fn()
		}

		loop = &aufs.Loop{
			Logger:  lagertest.NewTestLogger("test"),
			Retrier: fakeRetrier,
		}
	})

	AfterEach(func() {
		syscall.Unmount(destPath, 0)
		Expect(os.RemoveAll(destPath)).To(Succeed())
		Expect(os.Remove(bsFilePath)).To(Succeed())
	})

	Describe("MountFile", func() {
		It("mounts the file with noatime but not the journal_data_writeback options", func() {
			Expect(loop.MountFile(bsFilePath, destPath)).To(Succeed())

			session, err := gexec.Start(exec.Command("losetup", "-j", bsFilePath), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			loopDev := strings.TrimSpace(strings.Split(string(session.Out.Contents()), ":")[0])

			session, err = gexec.Start(exec.Command("cat", "/proc/mounts"), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say(fmt.Sprintf("%s %s ext4 rw,noatime", loopDev, destPath)))
			Expect(session).NotTo(gbytes.Say(",data=writeback"))
		})

		Context("when using a file that does not exist", func() {
			It("should return an error", func() {
				Expect(loop.MountFile("/path/to/my/nonexisting/banana", "/path/to/dest")).To(HaveOccurred())
			})
		})
	})

	Describe("Unmount", func() {
		It("should not leak devices", func() {
			var devicesAfterCreate, devicesAfterRelease int

			destPaths := make([]string, 10)
			for i := 0; i < 10; i++ {
				var err error

				tempFile, err := ioutil.TempFile("", "dinosaur")
				Expect(err).NotTo(HaveOccurred())

				_, err = exec.Command("truncate", "-s", "10M", tempFile.Name()).CombinedOutput()
				Expect(err).NotTo(HaveOccurred())
				_, err = exec.Command("mkfs.ext4", "-F", tempFile.Name()).CombinedOutput()
				Expect(err).NotTo(HaveOccurred())

				destPaths[i], err = ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(loop.MountFile(tempFile.Name(), destPaths[i])).To(Succeed())
			}

			output, err := exec.Command("sh", "-c", "losetup -a | grep dinosaur | wc -l").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			devicesAfterCreate, err = strconv.Atoi(strings.TrimSpace(string(output)))
			Expect(err).NotTo(HaveOccurred())

			for i := 0; i < 10; i++ {
				Expect(loop.Unmount(destPaths[i])).To(Succeed())
			}

			output, err = exec.Command("sh", "-c", "losetup -a | grep dinosaur | wc -l").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			devicesAfterRelease, err = strconv.Atoi(strings.TrimSpace(string(output)))
			Expect(err).NotTo(HaveOccurred())

			Expect(devicesAfterRelease).To(BeNumerically("~", devicesAfterCreate-10, 2))
		})

		Describe("retrying the unmount when it doesn't immediately work", func() {
			var testFile *os.File

			BeforeEach(func() {
				Expect(loop.MountFile(bsFilePath, destPath)).To(Succeed())
				var err error
				testFile, err = ioutil.TempFile(destPath, "")
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails when the unmount never succeeds", func() {
				defer func() {
					Expect(testFile.Close()).To(Succeed())
					Expect(loop.Unmount(destPath)).To(Succeed())
				}()

				Expect(loop.Unmount(destPath)).To(MatchError(ContainSubstring("unmounting file: exit status 1")))
			})

			It("suceeds when the unmount eventually succeeds", func() {
				fakeRetrier.RunStub = func(fn func() error) error {
					Expect(fn()).NotTo(Succeed())
					Expect(testFile.Close()).To(Succeed())
					Expect(fn()).To(Succeed())
					return nil
				}

				Expect(loop.Unmount(destPath)).To(Succeed())
			})
		})

		Context("when the provided mount point does not exist", func() {
			It("should succeed", func() {
				Expect(loop.Unmount("/dev/loopbanana")).To(Succeed())
			})
		})
	})
})
