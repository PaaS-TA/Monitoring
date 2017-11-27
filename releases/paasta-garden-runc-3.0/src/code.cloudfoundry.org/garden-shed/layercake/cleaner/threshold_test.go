package cleaner_test

import (
	"code.cloudfoundry.org/garden-shed/layercake/cleaner"
	"code.cloudfoundry.org/garden-shed/layercake/fake_cake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/image"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Threshold", func() {
	var (
		fakeCake  *fake_cake.FakeCake
		logger    lager.Logger
		threshold int64
	)

	BeforeEach(func() {
		threshold = 1234
		fakeCake = new(fake_cake.FakeCake)
		logger = lagertest.NewTestLogger("test")
	})

	Context("when there are no layers in the graph", func() {
		It("is not exceeded", func() {
			threshold := cleaner.NewThreshold(threshold)
			Expect(threshold.Exceeded(logger, fakeCake)).To(BeFalse())
		})
	})

	Context("when the limit is -1", func() {
		It("always returns false", func() {
			fakeCake.AllReturns([]*image.Image{{Size: 9999999}})
			threshold := cleaner.NewThreshold(-1)
			Expect(threshold.Exceeded(logger, fakeCake)).To(BeFalse())
		})
	})

	Context("when there is just one layer in the graph", func() {
		Context("and it exceeds the threshold", func() {
			BeforeEach(func() {
				fakeCake.AllReturns(
					[]*image.Image{
						&image.Image{
							Size: 1235,
						},
					},
				)
			})

			It("returns true", func() {
				threshold := cleaner.NewThreshold(threshold)
				Expect(threshold.Exceeded(logger, fakeCake)).To(BeTrue())
			})
		})

		Context("and it does not exceed the threshold", func() {
			BeforeEach(func() {
				fakeCake.AllReturns(
					[]*image.Image{
						&image.Image{
							Size: 1234,
						},
					},
				)
			})

			It("returns false", func() {
				threshold := cleaner.NewThreshold(threshold)
				Expect(threshold.Exceeded(logger, fakeCake)).To(BeFalse())
			})
		})
	})

	Context("when there are multiple layers", func() {
		Context("and it exceeds the threshold", func() {
			BeforeEach(func() {
				fakeCake.AllReturns(
					[]*image.Image{
						&image.Image{
							Size: 617,
						},
						&image.Image{
							Size: 618,
						},
					},
				)
			})

			It("returns true", func() {
				threshold := cleaner.NewThreshold(threshold)
				Expect(threshold.Exceeded(logger, fakeCake)).To(BeTrue())
			})
		})

		Context("and it does not exceed the threshold", func() {
			BeforeEach(func() {
				fakeCake.AllReturns(
					[]*image.Image{
						&image.Image{
							Size: 617,
						},
						&image.Image{
							Size: 617,
						},
					},
				)
			})

			It("returns true", func() {
				threshold := cleaner.NewThreshold(threshold)
				Expect(threshold.Exceeded(logger, fakeCake)).To(BeFalse())
			})
		})
	})
})
