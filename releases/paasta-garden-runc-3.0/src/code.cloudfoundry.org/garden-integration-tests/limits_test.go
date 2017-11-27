package garden_integration_tests_test

import (
	"os"

	"code.cloudfoundry.org/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Limits", func() {
	Describe("CPU limits", func() {
		BeforeEach(func() {
			limits.CPU = garden.CPULimits{
				LimitInShares: 100,
			}
		})

		It("reports the CPU limit", func() {
			cpuLimits, err := container.CurrentCPULimits()
			Expect(err).NotTo(HaveOccurred())
			Expect(cpuLimits.LimitInShares).To(BeEquivalentTo(100))
		})
	})

	Describe("memory limits", func() {
		BeforeEach(func() {
			limits = garden.Limits{Memory: garden.MemoryLimits{
				LimitInBytes: 64 * 1024 * 1024,
			}}
		})

		It("reports the memory limits", func() {
			memoryLimits, err := container.CurrentMemoryLimits()
			Expect(err).NotTo(HaveOccurred())
			Expect(memoryLimits.LimitInBytes).To(BeEquivalentTo(64 * 1024 * 1024))
		})

		It("kills a process if it uses too much memory", func() {
			process, err := container.Run(garden.ProcessSpec{
				User: "root",
				Path: "dd",
				Args: []string{"if=/dev/urandom", "of=/dev/shm/too-big", "bs=1M", "count=65"},
			}, garden.ProcessIO{})
			Expect(err).ToNot(HaveOccurred())

			Expect(process.Wait()).ToNot(Equal(0))
		})

		It("doesn't kill a process that uses lots of memory within the limit", func() {
			process, err := container.Run(garden.ProcessSpec{
				User: "root",
				Path: "dd",
				Args: []string{"if=/dev/urandom", "of=/dev/shm/almost-too-big", "bs=1M", "count=54"},
			}, ginkgoIO)
			Expect(err).ToNot(HaveOccurred())
			Expect(process.Wait()).To(Equal(0))
		})
	})

	Describe("disk limits", func() {
		BeforeEach(func() {
			if os.Getenv("EXTERNAL_IMAGE_PLUGIN_PROVIDED") == "true" {
				Skip("Disk limits are not enforced by garden when there is an external image plugin provided")
			}

			privilegedContainer = false

			limits.Disk.ByteSoft = 100 * 1024 * 1024
			limits.Disk.ByteHard = 100 * 1024 * 1024
			limits.Disk.Scope = garden.DiskLimitScopeTotal
		})

		DescribeTable("Metrics",
			func(reporter func() uint64) {
				createUser(container, "alice")

				initialBytes := reporter()
				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "dd",
					Args: []string{"if=/dev/zero", "of=/home/alice/some-file", "bs=1M", "count=3"},
				}, ginkgoIO)
				Expect(err).ToNot(HaveOccurred())
				Expect(process.Wait()).To(Equal(0))

				Eventually(reporter).Should(BeNumerically("~", initialBytes+3*1024*1024, 1024*1024))

				process, err = container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "dd",
					Args: []string{"if=/dev/zero", "of=/home/alice/another-file", "bs=1M", "count=10"},
				}, ginkgoIO)
				Expect(err).ToNot(HaveOccurred())
				Expect(process.Wait()).To(Equal(0))

				Eventually(reporter).Should(BeNumerically("~", initialBytes+uint64(13*1024*1024), 1024*1024))
			},

			Entry("with exclusive metrics", func() uint64 {
				metrics, err := container.Metrics()
				Expect(err).ToNot(HaveOccurred())
				return metrics.DiskStat.ExclusiveBytesUsed
			}),

			Entry("with total metrics", func() uint64 {
				metrics, err := container.Metrics()
				Expect(err).ToNot(HaveOccurred())
				return metrics.DiskStat.TotalBytesUsed
			}),
		)

		Context("when the scope is total", func() {
			BeforeEach(func() {
				imageRef.URI = "docker:///busybox#1.23"
				limits.Disk.ByteSoft = 10 * 1024 * 1024
				limits.Disk.ByteHard = 10 * 1024 * 1024
				limits.Disk.Scope = garden.DiskLimitScopeTotal
			})

			It("reports initial total bytes of a container based on size of image", func() {
				metrics, err := container.Metrics()
				Expect(err).ToNot(HaveOccurred())

				Expect(metrics.DiskStat.TotalBytesUsed).To(BeNumerically(">", metrics.DiskStat.ExclusiveBytesUsed))
				Expect(metrics.DiskStat.TotalBytesUsed).To(BeNumerically("~", 1024*1024, 512*1024)) // base busybox is > 1 MB but less than 1.5 MB
			})

			Context("and run a process that does not exceed the limit", func() {
				It("does not kill the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "root",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/root/test", "bs=1M", "count=7"},
					}, garden.ProcessIO{Stdout: GinkgoWriter, Stderr: GinkgoWriter})
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).To(Equal(0))
				})
			})

			Context("and run a process that exceeds the quota due to the size of the rootfs", func() {
				It("kills the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "root",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/root/test", "bs=1M", "count=9"}, // assume busybox itself accounts for > 1 MB
					}, garden.ProcessIO{Stdout: GinkgoWriter, Stderr: GinkgoWriter})
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})

			Context("when rootfs exceeds the quota", func() {
				BeforeEach(func() {
					assertContainerCreate = false
					imageRef.URI = "docker:///ubuntu#trusty-20160323"
				})

				It("should fail to create a container", func() {
					Expect(containerCreateErr).To(HaveOccurred())
				})
			})
		})

		Context("when the scope is exclusive", func() {
			BeforeEach(func() {
				limits.Disk.ByteSoft = 10 * 1024 * 1024
				limits.Disk.ByteHard = 10 * 1024 * 1024
				limits.Disk.Scope = garden.DiskLimitScopeExclusive
			})

			Context("and run a process that would exceed the quota due to the size of the rootfs", func() {
				It("does not kill the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "root",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/root/test", "bs=1M", "count=9"}, // should succeed, even though equivalent with 'total' scope does not
					}, garden.ProcessIO{Stdout: GinkgoWriter, Stderr: GinkgoWriter})
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).To(Equal(0))
				})
			})

			Context("and run a process that exceeds the quota", func() {
				It("kills the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "root",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/root/test", "bs=1M", "count=11"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})
		})

		Context("a rootfs with pre-existing users", func() {
			BeforeEach(func() {
				limits.Disk.ByteSoft = 10 * 1024 * 1024
				limits.Disk.ByteHard = 10 * 1024 * 1024
				limits.Disk.Scope = garden.DiskLimitScopeExclusive
			})

			JustBeforeEach(func() {
				createUser(container, "alice")
				createUser(container, "bob")
			})

			Context("and run a process that exceeds the quota as bob", func() {
				It("kills the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "bob",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/home/bob/test", "bs=1M", "count=11"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})

			Context("and run a process that exceeds the quota as alice", func() {
				It("kills the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/home/alice/test", "bs=1M", "count=11"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})

			Context("user alice is getting near the set limit", func() {
				JustBeforeEach(func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/home/alice/test", "bs=1M", "count=8"},
					}, garden.ProcessIO{
						Stderr: GinkgoWriter,
						Stdout: GinkgoWriter,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).To(Equal(0))
				})

				It("kills the process if user bob tries to exceed the shared limit", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "bob",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/home/bob/test", "bs=1M", "count=3"},
					}, garden.ProcessIO{
						Stderr: GinkgoWriter,
						Stdout: GinkgoWriter,
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})

			Context("when multiple containers are created for the same user", func() {
				var container2 garden.Container
				var err error

				BeforeEach(func() {
					limits.Disk.ByteSoft = 50 * 1024 * 1024
					limits.Disk.ByteHard = 50 * 1024 * 1024
					limits.Disk.Scope = garden.DiskLimitScopeExclusive
				})

				JustBeforeEach(func() {
					container2, err = gardenClient.Create(garden.ContainerSpec{
						Privileged: privilegedContainer,
						Image:      imageRef,
						Limits:     limits,
					})
					Expect(err).ToNot(HaveOccurred())

					createUser(container2, "alice")
				})

				AfterEach(func() {
					if container2 != nil {
						Expect(gardenClient.Destroy(container2.Handle())).To(Succeed())
					}
				})

				It("gives each container its own quota", func() {
					process, err := container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "dd",
						Args: []string{"if=/dev/urandom", "of=/tmp/some-file", "bs=1M", "count=40"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(process.Wait()).To(Equal(0))

					process, err = container2.Run(garden.ProcessSpec{
						User: "alice",
						Path: "dd",
						Args: []string{"if=/dev/urandom", "of=/tmp/some-file", "bs=1M", "count=40"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(process.Wait()).To(Equal(0))
				})
			})
		})

		Context("when the container is privileged", func() {
			BeforeEach(func() {
				privilegedContainer = true

				limits.Disk.ByteSoft = 10 * 1024 * 1024
				limits.Disk.ByteHard = 10 * 1024 * 1024
				limits.Disk.Scope = garden.DiskLimitScopeExclusive
			})

			Context("and run a process that exceeds the quota as root", func() {
				It("kills the process", func() {
					dd, err := container.Run(garden.ProcessSpec{
						User: "root",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/root/test", "bs=1M", "count=11"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})

			Context("and run a process that exceeds the quota as a new user", func() {
				It("kills the process", func() {
					createUser(container, "bob")

					dd, err := container.Run(garden.ProcessSpec{
						User: "bob",
						Path: "dd",
						Args: []string{"if=/dev/zero", "of=/home/bob/test", "bs=1M", "count=11"},
					}, ginkgoIO)
					Expect(err).ToNot(HaveOccurred())
					Expect(dd.Wait()).ToNot(Equal(0))
				})
			})
		})
	})

	Describe("PID limits", func() {
		Context("when there is a pid limit applied", func() {
			BeforeEach(func() {
				major, minor := getKernelVersion()
				if major < 4 || (major == 4 && minor < 4) {
					Skip("kernel version should be at 4.4 or later")
				}

				limits.Pid = garden.PidLimits{Max: 20}
			})

			It("prevents forking of processes", func() {
				stderr := gbytes.NewBuffer()
				process, err := container.Run(garden.ProcessSpec{
					User: "root",
					Path: "sh",
					Args: []string{"-c", "for i in `seq 1 20`; do sleep 2 & done"},
				}, garden.ProcessIO{
					Stdout: GinkgoWriter,
					Stderr: stderr,
				})
				Expect(err).NotTo(HaveOccurred())

				exitCode, err := process.Wait()
				Expect(exitCode).To(Equal(2))
				Expect(err).NotTo(HaveOccurred())
				Expect(stderr).To(gbytes.Say("sh: can't fork"))
			})
		})

		Context("when the pid limit is set to 0", func() {
			BeforeEach(func() {
				limits.Pid = garden.PidLimits{Max: 0}
			})

			It("applies no limit", func() {
				stderr := gbytes.NewBuffer()
				process, err := container.Run(garden.ProcessSpec{
					User: "root",
					Path: "sh",
					Args: []string{"-c", "ps"},
				}, garden.ProcessIO{
					Stdout: GinkgoWriter,
					Stderr: stderr,
				})
				Expect(err).NotTo(HaveOccurred())

				exitCode, err := process.Wait()
				Expect(exitCode).To(Equal(0))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
