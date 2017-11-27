package garden_integration_tests_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/garden"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	archiver "github.com/pivotal-golang/archiver/extractor/test_helper"
)

var _ = Describe("Lifecycle", func() {
	JustBeforeEach(func() {
		createUser(container, "alice")
	})

	Context("Creating a container with limits", func() {
		BeforeEach(func() {
			limits = garden.Limits{
				Memory: garden.MemoryLimits{
					LimitInBytes: 1024 * 1024 * 128,
				},
				CPU: garden.CPULimits{
					LimitInShares: 50,
				},
			}
		})

		It("it applies limits if set in the container spec", func() {
			memoryLimit, err := container.CurrentMemoryLimits()
			Expect(err).ToNot(HaveOccurred())
			Expect(memoryLimit).To(Equal(limits.Memory))

			cpuLimit, err := container.CurrentCPULimits()
			Expect(err).ToNot(HaveOccurred())
			Expect(cpuLimit).To(Equal(limits.CPU))
		})

		It("does not apply limits if not set in container spec", func() {
			diskLimit, err := container.CurrentDiskLimits()
			Expect(err).ToNot(HaveOccurred())
			Expect(diskLimit).To(Equal(garden.DiskLimits{}))

			bandwidthLimit, err := container.CurrentBandwidthLimits()
			Expect(err).ToNot(HaveOccurred())
			Expect(bandwidthLimit).To(Equal(garden.BandwidthLimits{}))
		})

		It("should be able to create and destroy containers sequentially", func() {
			if os.Getenv("NESTED") == "true" {
				Skip("Not supported on nested environments")
			}

			diskLimits := garden.DiskLimits{
				ByteHard: 2 * 1024 * 1024 * 1024,
			}

			container1, err := gardenClient.Create(garden.ContainerSpec{Limits: garden.Limits{Disk: diskLimits}})
			Expect(err).NotTo(HaveOccurred())
			container2, err := gardenClient.Create(garden.ContainerSpec{Limits: garden.Limits{Disk: diskLimits}})
			Expect(err).NotTo(HaveOccurred())

			Expect(gardenClient.Destroy(container1.Handle())).To(Succeed())
			Expect(gardenClient.Destroy(container2.Handle())).To(Succeed())
		})
	})

	Context("Creating a container with a duplicate handle", func() {
		It("returns a meaningful error message", func() {
			existingHandle := container.Handle()

			container, err := gardenClient.Create(garden.ContainerSpec{
				Handle: existingHandle,
			})

			Expect(container).To(BeNil())
			Expect(err).To(MatchError(fmt.Sprintf("Handle '%s' already in use", existingHandle)))
		})
	})

	It("returns garden.ContainerNotFound when destroying a container that doesn't exist", func() {
		Expect(gardenClient.Destroy("potato-sandwhich-policy")).To(MatchError(garden.ContainerNotFoundError{Handle: "potato-sandwhich-policy"}))
	})

	It("provides /dev/shm as tmpfs in the container", func() {
		process, err := container.Run(garden.ProcessSpec{
			User: "alice",
			Path: "dd",
			Args: []string{"if=/dev/urandom", "of=/dev/shm/some-data", "count=64", "bs=1k"},
		}, garden.ProcessIO{})
		Expect(err).ToNot(HaveOccurred())

		Expect(process.Wait()).To(Equal(0))

		outBuf := gbytes.NewBuffer()

		process, err = container.Run(garden.ProcessSpec{
			User: "alice",
			Path: "cat",
			Args: []string{"/proc/mounts"},
		}, garden.ProcessIO{
			Stdout: outBuf,
			Stderr: GinkgoWriter,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(process.Wait()).To(Equal(0))

		Expect(outBuf).To(gbytes.Say("tmpfs /dev/shm tmpfs"))
		Expect(outBuf).To(gbytes.Say("rw,nodev,relatime"))
	})

	It("gives the container a hostname based on its handle", func() {
		stdout := gbytes.NewBuffer()

		_, err := container.Run(garden.ProcessSpec{
			User: "alice",
			Path: "hostname",
		}, garden.ProcessIO{
			Stdout: stdout,
		})
		Expect(err).ToNot(HaveOccurred())

		Eventually(stdout).Should(gbytes.Say(fmt.Sprintf("%s\n", container.Handle())))
	})

	Context("when the handle is bigger than 49 characters", func() {
		BeforeEach(func() {
			handle = "7132-ec774112a9cd-101f8293-230e-4fa8-4138-e8244e6dcfa1"
		})

		It("should use the last 49 characters of the handle as the hostname", func() {
			stdout := gbytes.NewBuffer()

			_, err := container.Run(garden.ProcessSpec{
				User: "alice",
				Path: "hostname",
			}, garden.ProcessIO{
				Stdout: stdout,
			})
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("ec774112a9cd-101f8293-230e-4fa8-4138-e8244e6dcfa1"))
		})
	})

	Context("and sending a List request", func() {
		It("includes the created container", func() {
			Expect(getContainerHandles()).To(ContainElement(container.Handle()))
		})
	})

	Context("and sending an Info request", func() {
		It("returns the container's info", func() {
			info, err := container.Info()
			Expect(err).ToNot(HaveOccurred())

			Expect(info.State).To(Equal("active"))
		})
	})

	Describe("running a process", func() {
		Context("when root is requested", func() {
			It("runs as root inside the container", func() {
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

		It("streams output back and reports the exit status", func() {
			stdout := gbytes.NewBuffer()
			stderr := gbytes.NewBuffer()

			process, err := container.Run(garden.ProcessSpec{
				User: "alice",
				Path: "sh",
				Args: []string{"-c", "sleep 0.5; echo $FIRST; sleep 0.5; echo $SECOND >&2; sleep 0.5; exit 42"},
				Env:  []string{"FIRST=hello", "SECOND=goodbye"},
			}, garden.ProcessIO{
				Stdout: stdout,
				Stderr: stderr,
			})
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("hello\n"))
			Eventually(stderr).Should(gbytes.Say("goodbye\n"))
			Expect(process.Wait()).To(Equal(42))
		})

		Context("when multiple clients attach to the same process", func() {
			It("all clients attached should get the exit code", func() {
				process, err := container.Run(garden.ProcessSpec{
					Path: "sh",
					Args: []string{"-c", `sleep 2; exit 12`},
				}, garden.ProcessIO{})
				Expect(err).ToNot(HaveOccurred())

				wg := sync.WaitGroup{}
				for i := 0; i <= 5; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						defer GinkgoRecover()
						proc, err := container.Attach(process.ID(), garden.ProcessIO{})
						Expect(err).ToNot(HaveOccurred())
						code, err := proc.Wait()
						Expect(err).NotTo(HaveOccurred())
						Expect(code).To(Equal(12))
					}()
				}
				wg.Wait()
			})

			It("should be able to get the exitcode multiple times on the same process", func() {
				process, err := container.Run(garden.ProcessSpec{
					Path: "sh",
					Args: []string{"-c", `sleep 2; exit 12`},
				}, garden.ProcessIO{})
				Expect(err).ToNot(HaveOccurred())

				for i := 0; i < 3; i++ {
					code, err := process.Wait()
					Expect(err).ToNot(HaveOccurred())
					Expect(code).To(Equal(12))
				}
			})
		})

		It("all attached clients should get stdout and stderr", func() {
			var runStdout, attachStdout, runStderr, attachStderr bytes.Buffer

			process, err := container.Run(garden.ProcessSpec{
				Path: "sh",
				Args: []string{"-c", `sleep 1; for i in $(seq 1 10); do echo $i; echo $i >&2; done`},
			}, garden.ProcessIO{
				Stdout: io.MultiWriter(&runStdout, GinkgoWriter),
				Stderr: io.MultiWriter(&runStderr, GinkgoWriter),
			})
			Expect(err).ToNot(HaveOccurred())

			attachedProcess, err := container.Attach(process.ID(), garden.ProcessIO{
				Stdout: io.MultiWriter(&attachStdout, GinkgoWriter),
				Stderr: io.MultiWriter(&attachStderr, GinkgoWriter),
			})
			Expect(err).NotTo(HaveOccurred())

			exitCode, err := process.Wait()
			Expect(err).NotTo(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			// Looks redundant, but avoids race as we have 2 representations of the process
			exitCode, err = attachedProcess.Wait()
			Expect(err).NotTo(HaveOccurred())
			Expect(exitCode).To(Equal(0))

			Expect(runStdout.String()).To(Equal("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"), "1st buffer:")
			Expect(attachStdout.String()).To(Equal("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"), "2nd buffer:")

			Expect(runStderr.String()).To(Equal("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"), "1st buffer:")
			Expect(attachStderr.String()).To(Equal("1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n"), "2nd buffer:")
		})

		It("sends a TERM signal to the process if requested", func() {

			stdout := gbytes.NewBuffer()

			process, err := container.Run(garden.ProcessSpec{
				User: "alice",
				Path: "sh",
				Args: []string{"-c", `
				trap 'echo termed; exit 42' SIGTERM

				while true; do
					echo waiting
					sleep 1
				done
			`},
			}, garden.ProcessIO{
				Stdout: io.MultiWriter(GinkgoWriter, stdout),
				Stderr: GinkgoWriter,
			})
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("waiting"))
			Expect(process.Signal(garden.SignalTerminate)).To(Succeed())
			Eventually(stdout, "2s").Should(gbytes.Say("termed"))
			Expect(process.Wait()).To(Equal(42))
		})

		It("sends a TERM signal to the process run by root if requested", func() {

			stdout := gbytes.NewBuffer()

			process, err := container.Run(garden.ProcessSpec{
				User: "root",
				Path: "sh",
				Args: []string{"-c", `
				trap 'echo termed; exit 42' SIGTERM

				while true; do
					echo waiting
					sleep 1
				done
			`},
			}, garden.ProcessIO{
				Stdout: io.MultiWriter(GinkgoWriter, stdout),
				Stderr: GinkgoWriter,
			})
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("waiting"))
			Expect(process.Signal(garden.SignalTerminate)).To(Succeed())
			Eventually(stdout, "2s").Should(gbytes.Say("termed"))
			Expect(process.Wait()).To(Equal(42))
		})

		Context("even when /bin/kill does not exist", func() {
			JustBeforeEach(func() {
				process, err := container.Run(garden.ProcessSpec{
					User: "root",
					Path: "rm",
					Args: []string{"/bin/kill"},
				}, garden.ProcessIO{
					Stdout: GinkgoWriter,
					Stderr: GinkgoWriter,
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(process.Wait()).To(Equal(0))
			})

			checkProcessIsGone := func(container garden.Container, argsPrefix string) {
				Consistently(func() *gbytes.Buffer {
					stdout := gbytes.NewBuffer()
					process, err := container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "ps",
						Args: []string{"ax", "-o", "args="},
					}, garden.ProcessIO{
						Stdout: io.MultiWriter(stdout, GinkgoWriter),
						Stderr: GinkgoWriter,
					})
					Expect(err).ToNot(HaveOccurred())

					_, err = process.Wait()
					Expect(err).NotTo(HaveOccurred())

					return stdout
				}).ShouldNot(gbytes.Say(argsPrefix))
			}

			It("sends a KILL signal to the process if requested", func(done Done) {
				stdout := gbytes.NewBuffer()
				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{"-c", `
							trap wait SIGTERM

							while true; do
							  echo waiting
								sleep 1
							done
						`},
				}, garden.ProcessIO{
					Stdout: io.MultiWriter(GinkgoWriter, stdout),
					Stderr: GinkgoWriter,
				})
				Expect(err).ToNot(HaveOccurred())
				Eventually(stdout).Should(gbytes.Say("waiting"))

				Expect(process.Signal(garden.SignalKill)).To(Succeed())
				Expect(process.Wait()).To(Equal(137))

				checkProcessIsGone(container, "sh -c")

				close(done)
			}, 10.0)

			It("sends a TERMINATE signal to the process if requested", func(done Done) {
				stdout := gbytes.NewBuffer()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{"-c", `
							while true; do
							  echo waiting
								sleep 1
							done
						`},
				}, garden.ProcessIO{
					Stdout: io.MultiWriter(GinkgoWriter, stdout),
					Stderr: GinkgoWriter,
				})
				Expect(err).ToNot(HaveOccurred())
				Eventually(stdout).Should(gbytes.Say("waiting"))

				Expect(process.Signal(garden.SignalTerminate)).To(Succeed())
				Expect(process.Wait()).NotTo(BeZero())

				checkProcessIsGone(container, "sh -c while")

				close(done)
			}, 10.0)

			Context("when killing a process that does not use streaming", func() {
				var process garden.Process
				var buff *gbytes.Buffer

				JustBeforeEach(func() {
					var err error

					buff = gbytes.NewBuffer()
					process, err = container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "sh",
						Args: []string{
							"-c", "while true; do echo stillhere; sleep 1; done",
						},
					}, garden.ProcessIO{Stdout: buff})
					Expect(err).ToNot(HaveOccurred())

					Eventually(buff).Should(gbytes.Say("stillhere")) // make sure we dont kill before the process is spawned to avoid false-positives
					Expect(process.Signal(garden.SignalKill)).To(Succeed())
				})

				It("goes away", func(done Done) {
					Expect(process.Wait()).NotTo(Equal(0))
					Consistently(buff, "5s").ShouldNot(gbytes.Say("stillhere"))
					close(done)
				}, 30)
			})
		})

		It("avoids a race condition when sending a kill signal", func(done Done) {
			for i := 0; i < 20; i++ {
				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{"-c", `while true; do echo -n "x"; sleep 1; done`},
				}, garden.ProcessIO{
					Stdout: GinkgoWriter,
					Stderr: GinkgoWriter,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Signal(garden.SignalKill)).To(Succeed())
				Expect(process.Wait()).NotTo(Equal(0))
			}

			close(done)
		}, 480.0)

		It("collects the process's full output when tty is requested", func() {
			for i := 0; i < 100; i++ {
				stdout := gbytes.NewBuffer()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{"-c", "echo $(seq 10000)"},
					TTY:  &garden.TTYSpec{},
				}, garden.ProcessIO{
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))
				Expect(stdout).To(gbytes.Say("9999 10000"))
			}
		})

		It("collects the process's full output, even if it exits quickly after", func() {
			for i := 0; i < 100; i++ {
				stdout := gbytes.NewBuffer()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{"-c", "cat <&0"},
				}, garden.ProcessIO{
					Stdin:  bytes.NewBuffer([]byte("hi stdout")),
					Stderr: os.Stderr,
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))
				Expect(stdout).To(gbytes.Say("hi stdout"))
			}
		})

		It("streams input to the process's stdin", func() {
			stdout := gbytes.NewBuffer()

			process, err := container.Run(garden.ProcessSpec{
				User: "alice",
				Path: "sh",
				Args: []string{"-c", "cat <&0"},
			}, garden.ProcessIO{
				Stdin:  bytes.NewBufferString("hello\nworld"),
				Stdout: stdout,
			})
			Expect(err).ToNot(HaveOccurred())

			Eventually(stdout).Should(gbytes.Say("hello\nworld"))
			Expect(process.Wait()).To(Equal(0))
		})

		It("forwards the exit status even if stdin is still being written", func() {
			// this covers the case of intermediaries shuffling i/o around (e.g. wsh)
			// receiving SIGPIPE on write() due to the backing process exiting without
			// flushing stdin
			//
			// in practice it's flaky; sometimes write() finishes just before the
			// process exits, so run it ~10 times (observed it fail often in this range)

			for i := 0; i < 10; i++ {
				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "ls",
				}, garden.ProcessIO{
					Stdin: bytes.NewBufferString(strings.Repeat("x", 1024)),
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))
			}
		})

		Context("with a tty", func() {
			It("executes the process with a raw tty with the default window size", func() {
				stdout := gbytes.NewBuffer()
				_, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{
						"-c",
						`
						# The mechanism that is used to set TTY size (ioctl) is
						# asynchronous. Hence, stty does not return the correct result
						# right after the process is launched.
						while true; do
							stty -a
							sleep 1
						done
					`,
					},
					TTY: &garden.TTYSpec{},
				}, garden.ProcessIO{
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout, "3s").Should(gbytes.Say("rows 24; columns 80;"))
			})

			It("executes the process with a raw tty with the given window size", func() {
				stdout := gbytes.NewBuffer()
				_, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{
						"-c",
						`
						# The mechanism that is used to set TTY size (ioctl) is
						# asynchronous. Hence, stty does not return the correct result
						# right after the process is launched.
						while true; do
							stty -a
							sleep 1
						done
					`,
					},
					TTY: &garden.TTYSpec{
						WindowSize: &garden.WindowSize{
							Columns: 123,
							Rows:    456,
						},
					},
				}, garden.ProcessIO{
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout, "3s").Should(gbytes.Say("rows 456; columns 123;"))
			})

			It("executes the process with a raw tty and with onlcr to preserve formatting (\r\n, not just \n)", func() {
				stdout := gbytes.NewBuffer()
				_, err := container.Run(garden.ProcessSpec{
					Path: "sh",
					Args: []string{
						"-c",
						`
						while true; do
							echo -e "new\nline"
							sleep 1
					  done
					`,
					},
					TTY: &garden.TTYSpec{},
				}, garden.ProcessIO{
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("new\r\nline"))
			})

			It("can have its terminal resized", func() {
				stdout := gbytes.NewBuffer()

				inR, inW := io.Pipe()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{
						"-c",
						`
						trap "stty -a" SIGWINCH

						# continuously read so that the trap can keep firing
						while true; do
							echo waiting
							if read; then
								exit 0
							fi
						done
					`,
					},
					TTY: &garden.TTYSpec{
						WindowSize: &garden.WindowSize{
							Columns: 13,
							Rows:    46,
						},
					},
				}, garden.ProcessIO{
					Stdin:  inR,
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("waiting"))

				err = process.SetTTY(garden.TTYSpec{
					WindowSize: &garden.WindowSize{
						Columns: 123,
						Rows:    456,
					},
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("rows 456; columns 123;"))

				_, err = fmt.Fprintf(inW, "ok\n")
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))
			})

			It("all attached clients should get stdout and stderr", func() {
				var runStdout, attachStdout bytes.Buffer

				process, err := container.Run(garden.ProcessSpec{
					Path: "sh",
					Args: []string{"-c", `sleep 1; for i in $(seq 1 5); do echo $i; echo $i >&2; done`},
					TTY:  &garden.TTYSpec{},
				}, garden.ProcessIO{
					Stdout: io.MultiWriter(&runStdout, GinkgoWriter),
					Stderr: GinkgoWriter,
				})
				Expect(err).ToNot(HaveOccurred())

				attachedProcess, err := container.Attach(process.ID(), garden.ProcessIO{
					Stdout: io.MultiWriter(&attachStdout, GinkgoWriter),
					Stderr: GinkgoWriter,
				})
				Expect(err).NotTo(HaveOccurred())

				exitCode, err := process.Wait()
				Expect(err).NotTo(HaveOccurred())
				Expect(exitCode).To(Equal(0))

				// Looks redundant, but avoids race as we have 2 representations of the process
				exitCode, err = attachedProcess.Wait()
				Expect(err).NotTo(HaveOccurred())
				Expect(exitCode).To(Equal(0))

				Expect(runStdout.String()).To(Equal("1\r\n1\r\n2\r\n2\r\n3\r\n3\r\n4\r\n4\r\n5\r\n5\r\n"), "1st buffer:")
				Expect(attachStdout.String()).To(Equal("1\r\n1\r\n2\r\n2\r\n3\r\n3\r\n4\r\n4\r\n5\r\n5\r\n"), "2nd buffer:")
			})
		})

		Context("with a working directory", func() {
			It("executes with the working directory as the dir", func() {
				stdout := gbytes.NewBuffer()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "pwd",
					Dir:  "/usr",
				}, garden.ProcessIO{
					Stdout: stdout,
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout).Should(gbytes.Say("/usr\n"))
				Expect(process.Wait()).To(Equal(0))
			})
		})

		Context("and then sending a stop request", func() {
			It("terminates all running processes", func() {
				stdout := gbytes.NewBuffer()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{
						"-c",
						`
					trap 'exit 42' SIGTERM

					# sync with test, and allow trap to fire when not sleeping
					while true; do
						echo waiting
						sleep 1
					done
					`,
					},
				}, garden.ProcessIO{
					Stdout: stdout,
					Stderr: GinkgoWriter,
				})
				Expect(err).ToNot(HaveOccurred())

				Eventually(stdout, 30).Should(gbytes.Say("waiting"))

				err = container.Stop(false)
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(42))
			})

			It("recursively terminates all child processes", func(done Done) {
				defer close(done)

				stderr := gbytes.NewBuffer()

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "sh",
					Args: []string{
						"-c",
						`
					# don't die until child processes die
					trap wait SIGTERM

					# spawn child that exits when it receives TERM
					sh -c 'trap wait SIGTERM; sleep 100 & wait' &

					# sync with test. Use stderr to avoid buffering in the shell.
					echo waiting >&2

					# wait on children
					wait
					`,
					},
				}, garden.ProcessIO{
					Stderr: stderr,
				})

				Expect(err).ToNot(HaveOccurred())

				Eventually(stderr, 5).Should(gbytes.Say("waiting\n"))

				stoppedAt := time.Now()

				err = container.Stop(false)
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(143)) // 143 = 128 + SIGTERM

				Expect(time.Since(stoppedAt)).To(BeNumerically("<=", 9*time.Second))
			}, 15)

			It("changes the container's state to 'stopped'", func() {
				err := container.Stop(false)
				Expect(err).ToNot(HaveOccurred())

				info, err := container.Info()
				Expect(err).ToNot(HaveOccurred())

				Expect(info.State).To(Equal("stopped"))
			})

			Context("when a process does not die 10 seconds after receiving SIGTERM", func() {
				It("is forcibly killed", func() {
					stdout := gbytes.NewBuffer()
					process, err := container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "sh",
						Args: []string{
							"-c",
							`
							trap "echo cannot touch this" SIGTERM

							echo waiting
							while true
							do
								sleep 1000
							done
						`,
						},
					}, garden.ProcessIO{Stdout: stdout})

					Eventually(stdout).Should(gbytes.Say("waiting"))

					Expect(err).ToNot(HaveOccurred())

					stoppedAt := time.Now()

					err = container.Stop(false)
					Expect(err).ToNot(HaveOccurred())

					exitStatus, err := process.Wait()
					Expect(err).ToNot(HaveOccurred())
					if exitStatus != 137 && exitStatus != 255 {
						Fail(fmt.Sprintf("Unexpected exitStatus: %d", exitStatus))
					}

					Expect(time.Since(stoppedAt)).To(BeNumerically(">=", 10*time.Second))
				})
			})
		})

		Context("and streaming files in", func() {
			var tarStream io.Reader

			JustBeforeEach(func() {
				tmpdir, err := ioutil.TempDir("", "some-temp-dir-parent")
				Expect(err).ToNot(HaveOccurred())

				tgzPath := filepath.Join(tmpdir, "some.tgz")

				archiver.CreateTarGZArchive(
					tgzPath,
					[]archiver.ArchiveFile{
						{
							Name: "./some-temp-dir",
							Dir:  true,
						},
						{
							Name: "./some-temp-dir/some-temp-file",
							Body: "some-body",
						},
					},
				)

				tgz, err := os.Open(tgzPath)
				Expect(err).ToNot(HaveOccurred())

				tarStream, err = gzip.NewReader(tgz)
				Expect(err).ToNot(HaveOccurred())
			})

			It("creates the files in the container, as the specified user", func() {
				err := container.StreamIn(garden.StreamInSpec{
					User:      "alice",
					Path:      "/home/alice",
					TarStream: tarStream,
				})
				Expect(err).ToNot(HaveOccurred())

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "test",
					Args: []string{"-f", "/home/alice/some-temp-dir/some-temp-file"},
				}, garden.ProcessIO{})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))

				output := gbytes.NewBuffer()
				process, err = container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "ls",
					Args: []string{"-al", "/home/alice/some-temp-dir/some-temp-file"},
				}, garden.ProcessIO{
					Stdout: output,
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))

				// output should look like -rwxrwxrwx 1 alice alice 9 Jan  1  1970 /tmp/some-container-dir/some-temp-dir/some-temp-file
				Expect(output).To(gbytes.Say("alice"))
				Expect(output).To(gbytes.Say("alice"))
			})

			Context("when no user specified", func() {
				It("streams the files in as root", func() {
					err := container.StreamIn(garden.StreamInSpec{
						Path:      "/home/alice",
						TarStream: tarStream,
					})
					Expect(err).ToNot(HaveOccurred())

					out := gbytes.NewBuffer()
					process, err := container.Run(garden.ProcessSpec{
						User: "root",
						Path: "ls",
						Args: []string{"-la", "/home/alice/some-temp-dir/some-temp-file"},
					}, garden.ProcessIO{
						Stdout: out,
						Stderr: out,
					})
					Expect(err).ToNot(HaveOccurred())

					Expect(process.Wait()).To(Equal(0))
					Expect(string(out.Contents())).To(ContainSubstring("root"))
				})
			})

			Context("when a non-existent user specified", func() {
				It("returns error", func() {
					err := container.StreamIn(garden.StreamInSpec{
						User:      "batman",
						Path:      "/home/alice",
						TarStream: tarStream,
					})
					Expect(err).To(MatchError(ContainSubstring("error streaming in")))
				})
			})

			Context("when the specified user does not have permission to stream in", func() {
				JustBeforeEach(func() {
					createUser(container, "bob")
				})

				It("returns error", func() {
					err := container.StreamIn(garden.StreamInSpec{
						User:      "bob",
						Path:      "/home/alice",
						TarStream: tarStream,
					})
					Expect(err).To(MatchError(ContainSubstring("Permission denied")))
				})
			})

			Context("in a privileged container", func() {
				BeforeEach(func() {
					privilegedContainer = true
				})

				It("streams in relative to the default run directory", func() {
					err := container.StreamIn(garden.StreamInSpec{
						User:      "alice",
						Path:      ".",
						TarStream: tarStream,
					})
					Expect(err).ToNot(HaveOccurred())

					process, err := container.Run(garden.ProcessSpec{
						User: "alice",
						Path: "test",
						Args: []string{"-f", "some-temp-dir/some-temp-file"},
					}, garden.ProcessIO{})
					Expect(err).ToNot(HaveOccurred())

					Expect(process.Wait()).To(Equal(0))
				})
			})

			It("streams in relative to the default run directory", func() {
				err := container.StreamIn(garden.StreamInSpec{
					User:      "alice",
					Path:      ".",
					TarStream: tarStream,
				})
				Expect(err).ToNot(HaveOccurred())

				process, err := container.Run(garden.ProcessSpec{
					User: "alice",
					Path: "test",
					Args: []string{"-f", "some-temp-dir/some-temp-file"},
				}, garden.ProcessIO{})
				Expect(err).ToNot(HaveOccurred())

				Expect(process.Wait()).To(Equal(0))
			})

			It("returns an error when the tar process dies", func() {
				err := container.StreamIn(garden.StreamInSpec{
					User: "alice",
					Path: "/tmp/some-container-dir",
					TarStream: &io.LimitedReader{
						R: tarStream,
						N: 10,
					},
				})
				Expect(err).To(HaveOccurred())
			})

			Context("and then copying them out", func() {
				itStreamsTheDirectory := func(user string) {
					It("streams the directory", func() {
						process, err := container.Run(garden.ProcessSpec{
							User: "alice",
							Path: "sh",
							Args: []string{"-c", `mkdir -p some-outer-dir/some-inner-dir && touch some-outer-dir/some-inner-dir/some-file`},
						}, garden.ProcessIO{})
						Expect(err).ToNot(HaveOccurred())

						Expect(process.Wait()).To(Equal(0))

						tarOutput, err := container.StreamOut(garden.StreamOutSpec{
							User: user,
							Path: "/home/alice/some-outer-dir/some-inner-dir",
						})
						Expect(err).ToNot(HaveOccurred())

						tarReader := tar.NewReader(tarOutput)

						header, err := tarReader.Next()
						Expect(err).ToNot(HaveOccurred())
						Expect(header.Name).To(Equal("some-inner-dir/"))

						header, err = tarReader.Next()
						Expect(err).ToNot(HaveOccurred())
						Expect(header.Name).To(Equal("some-inner-dir/some-file"))
					})

				}

				itStreamsTheDirectory("alice")

				Context("when no user specified", func() {
					// Any user's files can be streamed out as root
					itStreamsTheDirectory("")
				})

				Context("with a trailing slash", func() {
					It("streams the contents of the directory", func() {
						process, err := container.Run(garden.ProcessSpec{
							User: "alice",
							Path: "sh",
							Args: []string{"-c", `mkdir -p some-container-dir && touch some-container-dir/some-file`},
						}, garden.ProcessIO{})
						Expect(err).ToNot(HaveOccurred())

						Expect(process.Wait()).To(Equal(0))

						tarOutput, err := container.StreamOut(garden.StreamOutSpec{
							User: "alice",
							Path: "some-container-dir/",
						})
						Expect(err).ToNot(HaveOccurred())

						tarReader := tar.NewReader(tarOutput)

						header, err := tarReader.Next()
						Expect(err).ToNot(HaveOccurred())
						Expect(header.Name).To(Equal("./"))

						header, err = tarReader.Next()
						Expect(err).ToNot(HaveOccurred())
						Expect(header.Name).To(Equal("./some-file"))
					})
				})
			})
		})
	})

	Context("when the container GraceTime is applied", func() {
		var containerHandle string

		It("should disappear after grace time and before timeout", func() {
			Expect(container.SetGraceTime(500 * time.Millisecond)).To(Succeed())
			containerHandle = container.Handle()

			_, err := gardenClient.Lookup(containerHandle)
			Expect(err).NotTo(HaveOccurred())
			container = nil // avoid double-destroying in AfterEach

			Eventually(func() error {
				_, err := gardenClient.Lookup(containerHandle)
				return err
			}, "10s", "1s").Should(HaveOccurred())
		})

		It("returns an unknown handle error when calling the API", func() {
			Eventually(func() error {
				return gardenClient.Destroy("not-a-real-handle")
			}).Should(MatchError(fmt.Sprintf("unknown handle: %s", "not-a-real-handle")))
		})
	})
})
