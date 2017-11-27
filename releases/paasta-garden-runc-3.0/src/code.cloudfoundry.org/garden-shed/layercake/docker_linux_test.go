package layercake_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"time"

	quotaedaufs "code.cloudfoundry.org/garden-shed/docker_drivers/aufs"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/graph"
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/archive"
	"github.com/eapache/go-resiliency/retrier"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	_ "github.com/docker/docker/daemon/graphdriver/aufs"
	_ "github.com/docker/docker/daemon/graphdriver/vfs"
	_ "github.com/docker/docker/pkg/chrootarchive" // allow reexec of docker-applyLayer
	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Init()
}

var _ = Describe("Docker", func() {
	var (
		root   string
		driver graphdriver.Driver
		cake   *layercake.Docker
	)

	BeforeEach(func() {
		var err error

		root, err = ioutil.TempDir("", "cakeroot")
		Expect(err).NotTo(HaveOccurred())

		Expect(syscall.Mount("tmpfs", root, "tmpfs", 0, "")).To(Succeed())
		driver, err := graphdriver.GetDriver("vfs", root, nil)
		Expect(err).NotTo(HaveOccurred())

		graph, err := graph.NewGraph(root, driver)
		Expect(err).NotTo(HaveOccurred())

		cake = &layercake.Docker{
			Graph:  graph,
			Driver: driver,
		}
	})

	AfterEach(func() {
		Expect(syscall.Unmount(root, 0)).To(Succeed())
		Expect(os.RemoveAll(root)).To(Succeed())
	})

	Describe("Register", func() {
		Context("after registering a layer", func() {
			var id layercake.ID
			var parent layercake.ID

			BeforeEach(func() {
				id = layercake.ContainerID("")
				parent = layercake.ContainerID("")
			})

			ItCanReadWriteTheLayer := func() {
				It("can read and write files", func() {
					p, err := cake.Path(id)
					Expect(err).NotTo(HaveOccurred())
					Expect(ioutil.WriteFile(path.Join(p, "foo"), []byte("hi"), 0700)).To(Succeed())

					p, err = cake.Path(id)
					Expect(err).NotTo(HaveOccurred())
					Expect(path.Join(p, "foo")).To(BeAnExistingFile())
				})

				It("can get back the image", func() {
					img, err := cake.Get(id)
					Expect(err).NotTo(HaveOccurred())
					Expect(img.ID).To(Equal(id.GraphID()))
					Expect(img.Parent).To(Equal(parent.GraphID()))
				})
			}

			Context("when the new layer is a docker image", func() {
				JustBeforeEach(func() {
					id = layercake.DockerImageID("70d8f0edf5c9008eb61c7c52c458e7e0a831649dbb238b93dde0854faae314a8")
					registerImageLayer(cake, &image.Image{
						ID:     id.GraphID(),
						Parent: parent.GraphID(),
					})
				})

				Context("without a parent", func() {
					ItCanReadWriteTheLayer()

					It("can read the files in the image", func() {
						p, err := cake.Path(id)
						Expect(err).NotTo(HaveOccurred())

						Expect(path.Join(p, id.GraphID())).To(BeAnExistingFile())
					})

					It("can be deleted", func() {
						cake.Remove(id)

						filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
							Expect(path).To(BeADirectory())
							return nil
						})
					})
				})

				Context("with a parent", func() {
					BeforeEach(func() {
						parent = layercake.DockerImageID("07d8fe0df5c9008eb16c7c52c548e7e0a831649dbb238b93dde0854faae3148a")
						registerImageLayer(cake, &image.Image{
							ID:     parent.GraphID(),
							Parent: "",
						})
					})

					ItCanReadWriteTheLayer()

					It("inherits files from the parent layer", func() {
						p, err := cake.Path(id)
						Expect(err).NotTo(HaveOccurred())

						Expect(path.Join(p, parent.GraphID())).To(BeAnExistingFile())
					})

					It("can read the files in the image", func() {
						p, err := cake.Path(id)
						Expect(err).NotTo(HaveOccurred())

						Expect(path.Join(p, id.GraphID())).To(BeAnExistingFile())
					})
				})
			})

			Context("when the new layer is a container", func() {
				Context("with a parent", func() {
					BeforeEach(func() {
						parent = layercake.DockerImageID("70d8f0edf5c9008eb61c7c52c458e7e0a831649dbb238b93dde0854faae314a8")
						registerImageLayer(cake, &image.Image{
							ID:     parent.GraphID(),
							Parent: "",
						})

						id = layercake.ContainerID("abc")
						createContainerLayer(cake, id, parent, "potato")
					})

					ItCanReadWriteTheLayer()

					It("inherits files from the parent layer", func() {
						p, err := cake.Path(id)
						Expect(err).NotTo(HaveOccurred())

						Expect(path.Join(p, parent.GraphID())).To(BeAnExistingFile())
					})

					It("saves the container ID in the graph", func() {
						p, err := cake.Get(id)
						Expect(err).NotTo(HaveOccurred())

						Expect(p.Container).To(Equal("potato"))
					})
				})
			})
		})
	})

	Describe("All", func() {
		BeforeEach(func() {
			createContainerLayer(cake, layercake.ContainerID("def"), layercake.DockerImageID(""), "")
			createContainerLayer(cake, layercake.ContainerID("abc"), layercake.ContainerID("def"), "")
			createContainerLayer(cake, layercake.ContainerID("child2"), layercake.ContainerID("def"), "")
		})

		It("returns all the layers in the graph", func() {
			Expect(cake.All()).To(HaveLen(3))

			var ids []string
			for _, layer := range cake.All() {
				ids = append(ids, layer.ID)
			}

			Expect(ids).To(ContainElement(layercake.ContainerID("def").GraphID()))
			Expect(ids).To(ContainElement(layercake.ContainerID("abc").GraphID()))
			Expect(ids).To(ContainElement(layercake.ContainerID("child2").GraphID()))
		})
	})

	Describe("IsLeaf", func() {
		BeforeEach(func() {
			createContainerLayer(cake, layercake.ContainerID("def"), layercake.DockerImageID(""), "")
			createContainerLayer(cake, layercake.ContainerID("abc"), layercake.ContainerID("def"), "")
			createContainerLayer(cake, layercake.ContainerID("child2"), layercake.ContainerID("def"), "")
		})

		Context("when an image has no children", func() {
			It("is a leaf", func() {
				Expect(cake.IsLeaf(layercake.ContainerID("abc"))).To(Equal(true))
			})
		})

		Context("when an image has children", func() {
			It("is not a leaf", func() {
				Expect(cake.IsLeaf(layercake.ContainerID("def"))).To(Equal(false))
			})
		})

		Context("when an image's final child is removed", func() {
			It("is becomes a leaf", func() {
				Expect(cake.IsLeaf(layercake.ContainerID("def"))).To(Equal(false))

				Expect(cake.Remove(layercake.ContainerID("abc"))).To(Succeed())
				Expect(cake.IsLeaf(layercake.ContainerID("def"))).To(Equal(false))

				Expect(cake.Remove(layercake.ContainerID("child2"))).To(Succeed())
				Expect(cake.IsLeaf(layercake.ContainerID("def"))).To(Equal(true))
			})
		})
	})

	Describe("GetAllLeaves", func() {
		BeforeEach(func() {
			createContainerLayer(cake, layercake.ContainerID("def"), layercake.DockerImageID(""), "")
			createContainerLayer(cake, layercake.ContainerID("abc"), layercake.ContainerID("def"), "")
			createContainerLayer(cake, layercake.ContainerID("child2"), layercake.ContainerID("def"), "")
		})

		It("should return all the leaves", func() {
			leaves, err := cake.GetAllLeaves()
			Expect(err).NotTo(HaveOccurred())

			Expect(leaves).To(HaveLen(2))
			Expect(leaves).To(ContainElement(layercake.DockerImageID(layercake.ContainerID("abc").GraphID())))
			Expect(leaves).To(ContainElement(layercake.DockerImageID(layercake.ContainerID("child2").GraphID())))
		})
	})

	Describe("QuotaedPath", func() {
		Context("when not using aufs quotaed driver", func() {
			It("should return an error", func() {
				id := layercake.ContainerID("aubergine-layer")

				registerImageLayer(cake, &image.Image{
					ID: id.GraphID(),
				})

				_, err := cake.QuotaedPath(id, 10*1024*1024)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when using aufs quotaed driver", func() {
			var (
				backingStoreRoot string
				id               layercake.ContainerID
			)

			BeforeEach(func() {
				var err error

				backingStoreRoot, err = ioutil.TempDir("", "backingstore")
				Expect(err).NotTo(HaveOccurred())

				Expect(syscall.Mount("tmpfs", root, "tmpfs", 0, "")).To(Succeed())

				driver, err = graphdriver.GetDriver("aufs", root, nil)
				Expect(err).NotTo(HaveOccurred())

				driver = &quotaedaufs.QuotaedDriver{
					GraphDriver: driver,
					Unmount:     quotaedaufs.Unmount,
					BackingStoreMgr: &quotaedaufs.BackingStore{
						RootPath: backingStoreRoot,
						Logger:   lagertest.NewTestLogger("test"),
					},
					LoopMounter: &quotaedaufs.Loop{
						Retrier: retrier.New(retrier.ConstantBackoff(200, 500*time.Millisecond), nil),
						Logger:  lagertest.NewTestLogger("test"),
					},
					Retrier:  retrier.New(retrier.ConstantBackoff(200, 500*time.Millisecond), nil),
					RootPath: root,
					Logger:   lagertest.NewTestLogger("test"),
				}

				graph, err := graph.NewGraph(root, driver)
				Expect(err).NotTo(HaveOccurred())

				cake = &layercake.Docker{
					Graph:  graph,
					Driver: driver,
				}
			})

			JustBeforeEach(func() {
				parent := layercake.DockerImageID("70d8f0edf5c9008eb61c7c52c458e7e0a831649dbb238b93dde0854faae314a8")
				registerImageLayer(cake, &image.Image{
					ID:     parent.GraphID(),
					Parent: "",
				})

				id = layercake.ContainerID("abc")
				createContainerLayer(cake, id, parent, "")
			})

			AfterEach(func() {
				Expect(cake.Remove(id)).To(Succeed())
				Expect(driver.Cleanup()).To(Succeed())
				Expect(syscall.Unmount(root, 0)).To(Succeed())
				Expect(os.RemoveAll(backingStoreRoot)).To(Succeed())
			})

			It("should allow the user to write a file that does not exceed the quota", func() {
				path, err := cake.QuotaedPath(id, 10*1024*1024)
				Expect(err).NotTo(HaveOccurred())

				cmd := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", filepath.Join(path, "foo")), "bs=1M", "count=8")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).To(Succeed())

				Eventually(session).Should(gexec.Exit(0))
			})

			It("should not allow the user to write that exceeds the quota", func() {
				path, err := cake.QuotaedPath(id, 10*1024*1024)
				Expect(err).NotTo(HaveOccurred())

				cmd := exec.Command("dd", "if=/dev/zero", fmt.Sprintf("of=%s", filepath.Join(path, "foo")), "bs=1M", "count=11")
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).To(Succeed())

				Eventually(session).ShouldNot(gexec.Exit(0))
			})
		})
	})
})

func createContainerLayer(cake *layercake.Docker, id, parent layercake.ID, containerId string) {
	Expect(cake.Create(id, parent, containerId)).To(Succeed())
}

func registerImageLayer(cake *layercake.Docker, img *image.Image) {
	tmp, err := ioutil.TempDir("", "my-img")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(tmp)

	Expect(ioutil.WriteFile(path.Join(tmp, img.ID), []byte("Hello"), 0700)).To(Succeed())
	archiver, _ := archive.Tar(tmp, archive.Uncompressed)

	Expect(cake.Register(img, archiver)).To(Succeed())
}
