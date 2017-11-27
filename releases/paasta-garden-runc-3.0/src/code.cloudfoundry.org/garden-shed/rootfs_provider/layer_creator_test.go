package rootfs_provider_test

import (
	"errors"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/garden-shed/layercake/fake_cake"
	"code.cloudfoundry.org/garden-shed/repository_fetcher"
	"code.cloudfoundry.org/garden-shed/rootfs_provider"
	"code.cloudfoundry.org/garden-shed/rootfs_provider/fake_namespacer"
	"code.cloudfoundry.org/garden-shed/rootfs_spec"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeVolumeCreator struct {
	Created     []RootAndVolume
	CreateError error
}

type RootAndVolume struct {
	RootPath string
	Volume   string
}

func (f *FakeVolumeCreator) Create(path, v string) error {
	f.Created = append(f.Created, RootAndVolume{path, v})
	return f.CreateError
}

var _ = Describe("Layer Creator", func() {
	var (
		fakeCake          *fake_cake.FakeCake
		fakeNamespacer    *fake_namespacer.FakeNamespacer
		fakeVolumeCreator *FakeVolumeCreator
		name              string

		provider *rootfs_provider.ContainerLayerCreator
	)

	BeforeEach(func() {
		fakeCake = new(fake_cake.FakeCake)
		fakeVolumeCreator = &FakeVolumeCreator{}
		fakeNamespacer = &fake_namespacer.FakeNamespacer{}
		name = "some-name"

		provider = rootfs_provider.NewLayerCreator(
			fakeCake,
			fakeVolumeCreator,
			fakeNamespacer,
		)
	})

	Describe("Create", func() {
		Context("when the namespace parameter is false", func() {
			It("creates a graph entry with it as the parent", func() {
				fakeCake.PathReturns("/some/graph/driver/mount/point", nil)

				mountpoint, envvars, err := provider.Create(
					lagertest.NewTestLogger("test"),
					"some-id",
					&repository_fetcher.Image{
						ImageID: "some-image-id",
						Env:     []string{"env1=env1value", "env2=env2value"},
					},
					rootfs_spec.Spec{
						Namespaced: false,
						QuotaSize:  0,
					},
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeCake.CreateCallCount()).To(Equal(1))
				id, parent, containerID := fakeCake.CreateArgsForCall(0)
				Expect(id).To(Equal(layercake.ContainerID("some-id")))
				Expect(parent).To(Equal(layercake.DockerImageID("some-image-id")))
				Expect(containerID).To(Equal("some-id"))

				Expect(mountpoint).To(Equal("/some/graph/driver/mount/point"))
				Expect(envvars).To(Equal(
					[]string{
						"env1=env1value",
						"env2=env2value",
					},
				))
			})
		})

		Context("when the quota is positive", func() {
			It("should return the quotaed mount point path", func() {
				quotaedPath := "/path/to/quotaed/bananas"

				fakeCake.QuotaedPathReturns(quotaedPath, nil)

				mountPointPath, _, err := provider.Create(
					lagertest.NewTestLogger("test"),
					"some-id",
					&repository_fetcher.Image{
						ImageID: "some-image-id",
					},
					rootfs_spec.Spec{
						Namespaced: false,
						QuotaSize:  int64(10 * 1024 * 1024),
					},
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(mountPointPath).To(Equal(quotaedPath))
			})

			Context("and the scope is exclusive", func() {
				It("should get the quotaed path", func() {
					id := "some-id"
					quota := int64(10 * 1024 * 1024)

					_, _, err := provider.Create(
						lagertest.NewTestLogger("test"),
						id,
						&repository_fetcher.Image{
							ImageID: "some-image-id",
						},
						rootfs_spec.Spec{
							Namespaced: false,
							QuotaSize:  quota,
							QuotaScope: garden.DiskLimitScopeExclusive,
						},
					)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCake.QuotaedPathCallCount()).To(Equal(1))
					reqId, reqQuota := fakeCake.QuotaedPathArgsForCall(0)
					Expect(reqQuota).To(Equal(quota))
					Expect(reqId).To(Equal(layercake.ContainerID(id)))
				})
			})

			Context("and the scope is total", func() {
				It("should get the quotaed path", func() {
					id := "some-id"
					quota := int64(10 * 1024 * 1024)
					imageSize := int64(5 * 1024 * 1024)

					_, _, err := provider.Create(
						lagertest.NewTestLogger("test"),
						id,
						&repository_fetcher.Image{
							ImageID: "some-image-id",
							Size:    imageSize,
						},
						rootfs_spec.Spec{
							Namespaced: false,
							QuotaSize:  quota,
							QuotaScope: garden.DiskLimitScopeTotal,
						},
					)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCake.QuotaedPathCallCount()).To(Equal(1))
					reqId, reqQuota := fakeCake.QuotaedPathArgsForCall(0)
					Expect(reqQuota).To(Equal(quota - imageSize))
					Expect(reqId).To(Equal(layercake.ContainerID(id)))
				})
			})

			Context("when the layer cake fails to mount the quotaed volume", func() {
				BeforeEach(func() {
					fakeCake.QuotaedPathReturns("", errors.New("my banana tastes weird"))
				})

				It("should return an error", func() {
					_, _, err := provider.Create(
						lagertest.NewTestLogger("test"),
						"some-id",
						&repository_fetcher.Image{
							ImageID: "some-image-id",
						},
						rootfs_spec.Spec{
							Namespaced: false,
							QuotaSize:  10 * 1024 * 1024,
						},
					)
					Expect(err).To(MatchError(ContainSubstring("my banana tastes weird")))
				})

				It("should not create the volumes", func() {
					_, _, err := provider.Create(
						lagertest.NewTestLogger("test"),
						"some-id",
						&repository_fetcher.Image{
							ImageID: "some-image-id",
							Volumes: []string{"/foo", "/bar"},
						},
						rootfs_spec.Spec{
							Namespaced: false,
							QuotaSize:  10 * 1024 * 1024,
						},
					)
					Expect(err).To(HaveOccurred())

					Expect(fakeVolumeCreator.Created).To(BeEmpty())
				})
			})
		})

		Context("when the namespace parameter is true", func() {
			Context("and the image has not been translated yet", func() {
				var (
					mountpoint string
					envvars    []string
				)

				JustBeforeEach(func() {
					fakeCake.GetReturns(nil, errors.New("no image here"))

					fakeCake.PathStub = func(id layercake.ID) (string, error) {
						return "/mount/point/" + id.GraphID(), nil
					}

					fakeNamespacer.CacheKeyReturns("jam")

					var err error
					mountpoint, envvars, err = provider.Create(
						lagertest.NewTestLogger("test"),
						"some-id",
						&repository_fetcher.Image{
							ImageID: "some-image-id",
							Env:     []string{"env1=env1value", "env2=env2value"},
						},
						rootfs_spec.Spec{
							Namespaced: true,
							QuotaSize:  0,
						},
					)

					Expect(err).ToNot(HaveOccurred())
				})

				It("namespaces it, and creates a graph entry with it as the parent", func() {
					Expect(fakeCake.CreateCallCount()).To(Equal(2))
					id, parent, _ := fakeCake.CreateArgsForCall(0)
					Expect(id).To(Equal(layercake.NamespacedID(layercake.DockerImageID("some-image-id"), "jam")))
					Expect(parent).To(Equal(layercake.DockerImageID("some-image-id")))

					id, parent, containerID := fakeCake.CreateArgsForCall(1)
					Expect(id).To(Equal(layercake.ContainerID("some-id")))
					Expect(parent).To(Equal(layercake.NamespacedID(layercake.DockerImageID("some-image-id"), "jam")))
					Expect(containerID).To(Equal("some-id"))

					Expect(fakeNamespacer.NamespaceCallCount()).To(Equal(1))
					_, dst := fakeNamespacer.NamespaceArgsForCall(0)
					Expect(dst).To(Equal("/mount/point/" + layercake.NamespacedID(layercake.DockerImageID("some-image-id"), "jam").GraphID()))

					Expect(mountpoint).To(Equal("/mount/point/" + layercake.ContainerID("some-id").GraphID()))
					Expect(envvars).To(Equal(
						[]string{
							"env1=env1value",
							"env2=env2value",
						},
					))
				})

				Context("unmounting the translation layer", func() {
					BeforeEach(func() {
						// ensure umount doesnt happen too quickly
						fakeNamespacer.NamespaceStub = func(_ lager.Logger, _ string) error {
							Expect(fakeCake.UnmountCallCount()).To(Equal(0))
							return nil
						}
					})

					It("unmounts the translation layer after performing namespacing", func() {
						Expect(fakeCake.UnmountCallCount()).Should(Equal(1))
						Expect(fakeCake.UnmountArgsForCall(0)).To(Equal(layercake.NamespacedID(layercake.DockerImageID("some-image-id"), "jam")))
					})
				})
			})

			Context("and the image has already been translated", func() {
				BeforeEach(func() {
					fakeCake.PathStub = func(id layercake.ID) (string, error) {
						return "/mount/point/" + id.GraphID(), nil
					}

					fakeNamespacer.CacheKeyReturns("sandwich")

					fakeCake.GetStub = func(id layercake.ID) (*image.Image, error) {
						if id == (layercake.NamespacedID(layercake.DockerImageID("some-image-id"), "sandwich")) {
							return &image.Image{}, nil
						}

						return nil, errors.New("hello")
					}

				})

				It("reuses the translated layer", func() {
					mountpoint, envvars, err := provider.Create(
						lagertest.NewTestLogger("test"),
						"some-id",
						&repository_fetcher.Image{
							ImageID: "some-image-id",
							Env:     []string{"env1=env1value", "env2=env2value"},
						},
						rootfs_spec.Spec{
							Namespaced: true,
							QuotaSize:  0,
						},
					)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeCake.CreateCallCount()).To(Equal(1))
					id, parent, _ := fakeCake.CreateArgsForCall(0)
					Expect(id).To(Equal(layercake.ContainerID("some-id")))
					Expect(parent).To(Equal(layercake.NamespacedID(layercake.DockerImageID("some-image-id"), "sandwich")))

					Expect(fakeNamespacer.NamespaceCallCount()).To(Equal(0))

					Expect(mountpoint).To(Equal("/mount/point/" + layercake.ContainerID("some-id").GraphID()))
					Expect(envvars).To(Equal(
						[]string{
							"env1=env1value",
							"env2=env2value",
						},
					))
				})
			})
		})

		Context("when the image has associated VOLUMEs", func() {
			It("creates empty directories for all volumes", func() {
				fakeCake.PathReturns("/some/graph/driver/mount/point", nil)

				_, _, err := provider.Create(
					lagertest.NewTestLogger("test"),
					"some-id",
					&repository_fetcher.Image{ImageID: "some-image-id", Volumes: []string{"/foo", "/bar"}},
					rootfs_spec.Spec{
						Namespaced: false,
						QuotaSize:  0,
					},
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeVolumeCreator.Created).To(Equal(
					[]RootAndVolume{
						{"/some/graph/driver/mount/point", "/foo"},
						{"/some/graph/driver/mount/point", "/bar"},
					}))
			})

			Context("when creating a volume fails", func() {
				It("returns an error", func() {
					fakeCake.PathReturns("/some/graph/driver/mount/point", nil)
					fakeVolumeCreator.CreateError = errors.New("o nooo")

					_, _, err := provider.Create(
						lagertest.NewTestLogger("test"),
						"some-id",
						&repository_fetcher.Image{ImageID: "some-image-id", Volumes: []string{"/foo", "/bar"}},
						rootfs_spec.Spec{
							Namespaced: false,
							QuotaSize:  0,
						},
					)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("but creating the graph entry fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeCake.CreateReturns(disaster)
			})

			It("returns the error", func() {
				_, _, err := provider.Create(
					lagertest.NewTestLogger("test"),
					"some-id",
					&repository_fetcher.Image{ImageID: "some-image-id"},
					rootfs_spec.Spec{
						Namespaced: false,
						QuotaSize:  0,
					},
				)
				Expect(err).To(Equal(disaster))
			})
		})

		Context("but getting the graph entry fails", func() {
			disaster := errors.New("oh no!")

			BeforeEach(func() {
				fakeCake.PathReturns("", disaster)
			})

			It("returns the error", func() {
				_, _, err := provider.Create(
					lagertest.NewTestLogger("test"),
					"some-id",
					&repository_fetcher.Image{ImageID: "some-image-id"},
					rootfs_spec.Spec{
						Namespaced: false,
						QuotaSize:  0,
					},
				)
				Expect(err).To(Equal(disaster))
			})
		})
	})
})
