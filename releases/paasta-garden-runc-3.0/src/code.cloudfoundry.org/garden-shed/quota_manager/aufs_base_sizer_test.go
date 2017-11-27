package quota_manager_test

import (
	"errors"

	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/garden-shed/layercake/fake_cake"
	"code.cloudfoundry.org/garden-shed/quota_manager"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/image"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AUFSBaseSizer", func() {
	Describe("BaseSize", func() {
		var (
			baseSizer *quota_manager.AUFSBaseSizer
			fakeCake  *fake_cake.FakeCake

			logger lager.Logger
		)

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("test")
			fakeCake = new(fake_cake.FakeCake)
			fakeCake.GetReturns(nil, errors.New("no such image"))

			baseSizer = quota_manager.NewAUFSBaseSizer(fakeCake)
		})

		It("asks for the size of the layer based on the base name of the rootfs path", func() {
			baseSizer.BaseSize(logger, "/some/path/to/54321")
			Expect(fakeCake.GetCallCount()).To(Equal(1))
			Expect(fakeCake.GetArgsForCall(0)).To(Equal(layercake.DockerImageID("54321")))
		})

		Context("when the layer doesn't exist", func() {
			It("returns an error", func() {
				_, err := baseSizer.BaseSize(logger, "/i/dont/exist")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the layer exists", func() {
			Context("and has no parents", func() {
				It("returns the size of the layer", func() {
					fakeCake.GetReturns(&image.Image{
						Size: 1234,
					}, nil)

					size, err := baseSizer.BaseSize(logger, "/i/do/exist")
					Expect(err).NotTo(HaveOccurred())
					Expect(size).To(BeNumerically("==", 1234))
				})
			})

			Context("and has parents", func() {
				It("returns the total size of all the parents", func() {
					imgs := map[string]*image.Image{
						"child":   &image.Image{Parent: "parent1", Size: 1234},
						"parent1": &image.Image{Parent: "parent2", Size: 456},
						"parent2": &image.Image{Size: 789},
					}

					fakeCake.GetStub = func(id layercake.ID) (*image.Image, error) {
						return imgs[id.GraphID()], nil
					}

					size, err := baseSizer.BaseSize(logger, "/i/have/parent/layers/child")
					Expect(err).NotTo(HaveOccurred())
					Expect(size).To(BeNumerically("==", 1234+456+789))
				})
			})
		})
	})
})
