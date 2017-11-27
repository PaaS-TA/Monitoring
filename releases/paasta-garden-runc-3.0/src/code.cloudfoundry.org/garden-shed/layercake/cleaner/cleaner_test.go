package cleaner_test

import (
	"errors"

	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/garden-shed/layercake/cleaner"
	fakes "code.cloudfoundry.org/garden-shed/layercake/cleaner/cleanerfakes"
	"code.cloudfoundry.org/garden-shed/layercake/fake_cake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/image"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Oven cleaner", func() {
	var (
		retainer      cleaner.RetainChecker
		gc            *cleaner.OvenCleaner
		fakeCake      *fake_cake.FakeCake
		child2parent  map[layercake.ID]layercake.ID // child -> parent
		size          map[layercake.ID]int64
		fakeThreshold *fakes.FakeThreshold
		logger        lager.Logger
	)

	BeforeEach(func() {
		fakeThreshold = new(fakes.FakeThreshold)
		logger = lagertest.NewTestLogger("test")

		retainer = cleaner.NewRetainer()

		fakeCake = new(fake_cake.FakeCake)
		fakeCake.GetStub = func(id layercake.ID) (*image.Image, error) {
			if parent, ok := child2parent[id]; ok {
				return &image.Image{
					ID:     id.GraphID(),
					Parent: parent.GraphID(),
					Size:   size[id],
				}, nil
			}

			return &image.Image{
				Size: size[id],
			}, nil
		}

		fakeCake.IsLeafStub = func(id layercake.ID) (bool, error) {
			for _, p := range child2parent {
				if p == id {
					return false, nil
				}
			}

			return true, nil
		}

		fakeCake.RemoveStub = func(id layercake.ID) error {
			delete(child2parent, id)
			return nil
		}

		child2parent = make(map[layercake.ID]layercake.ID)
		size = make(map[layercake.ID]int64)
	})

	JustBeforeEach(func() {
		gc = cleaner.NewOvenCleaner(
			retainer,
			fakeThreshold,
		)
	})

	Context("when the threshold is exceeded", func() {
		BeforeEach(func() {
			fakeThreshold.ExceededReturns(true)
		})

		Describe("GC", func() {
			Context("when there is a single leaf", func() {
				BeforeEach(func() {
					fakeCake.GetAllLeavesReturns([]layercake.ID{layercake.DockerImageID("child")}, nil)
					size[layercake.DockerImageID("child2")] = 2048
				})

				It("should not remove it when it is used by a container", func() {
					fakeCake.GetReturns(&image.Image{Container: "used-by-me"}, nil)
					Expect(gc.GC(logger, fakeCake)).To(Succeed())
					Expect(fakeCake.RemoveCallCount()).To(Equal(0))
				})

				Context("when the layer has no parents", func() {
					BeforeEach(func() {
						fakeCake.GetReturns(&image.Image{}, nil)
					})

					It("removes the layer", func() {
						Expect(gc.GC(logger, fakeCake)).To(Succeed())
						Expect(fakeCake.RemoveCallCount()).To(Equal(1))
						Expect(fakeCake.RemoveArgsForCall(0)).To(Equal(layercake.DockerImageID("child")))
					})

					Context("when the layer is retained", func() {
						JustBeforeEach(func() {
							retainer.Retain(lagertest.NewTestLogger(""), layercake.DockerImageID("child"))
						})

						It("should not remove the layer", func() {
							Expect(gc.GC(logger, fakeCake)).To(Succeed())
							Expect(fakeCake.RemoveCallCount()).To(Equal(0))
						})
					})

					Context("when removing fails", func() {
						It("returns an error", func() {
							fakeCake.RemoveReturns(errors.New("cake failure"))
							Expect(gc.GC(logger, fakeCake)).To(MatchError("cake failure"))
						})
					})
				})

				Context("when the layer has a parent", func() {
					BeforeEach(func() {
						child2parent[layercake.DockerImageID("child")] = layercake.DockerImageID("parent")
					})

					Context("and the parent has no other children", func() {
						It("removes the layer, and its parent", func() {
							Expect(gc.GC(logger, fakeCake)).To(Succeed())

							Expect(fakeCake.RemoveCallCount()).To(Equal(2))
							Expect(fakeCake.RemoveArgsForCall(0)).To(Equal(layercake.DockerImageID("child")))
							Expect(fakeCake.RemoveArgsForCall(1)).To(Equal(layercake.DockerImageID("parent")))
						})
					})

					Context("when removing fails", func() {
						It("does not remove any more layers", func() {
							fakeCake.RemoveReturns(errors.New("cake failure"))
							gc.GC(logger, fakeCake)
							Expect(fakeCake.RemoveCallCount()).To(Equal(1))
						})
					})

					Context("but the layer has another child", func() {
						BeforeEach(func() {
							child2parent[layercake.DockerImageID("some-other-child")] = layercake.DockerImageID("parent")
						})

						It("removes only the initial layer", func() {
							child2parent[layercake.DockerImageID("child")] = layercake.DockerImageID("parent")
							Expect(gc.GC(logger, fakeCake)).To(Succeed())

							Expect(fakeCake.RemoveCallCount()).To(Equal(1))
							Expect(fakeCake.RemoveArgsForCall(0)).To(Equal(layercake.DockerImageID("child")))
						})
					})
				})

				Context("when the layer has grandparents", func() {
					It("removes all the grandparents", func() {
						child2parent[layercake.DockerImageID("child")] = layercake.DockerImageID("parent")
						child2parent[layercake.DockerImageID("parent")] = layercake.DockerImageID("granddaddy")

						Expect(gc.GC(logger, fakeCake)).To(Succeed())

						Expect(fakeCake.RemoveCallCount()).To(Equal(3))
						Expect(fakeCake.RemoveArgsForCall(0)).To(Equal(layercake.DockerImageID("child")))
						Expect(fakeCake.RemoveArgsForCall(1)).To(Equal(layercake.DockerImageID("parent")))
						Expect(fakeCake.RemoveArgsForCall(2)).To(Equal(layercake.DockerImageID("granddaddy")))
					})
				})
			})

			Context("when there are multiple leaves", func() {
				BeforeEach(func() {
					fakeCake.GetAllLeavesReturns([]layercake.ID{layercake.DockerImageID("child1"), layercake.DockerImageID("child2")}, nil)
				})

				It("removes all of the leaves", func() {
					Expect(gc.GC(logger, fakeCake)).To(Succeed())
					Expect(fakeCake.RemoveCallCount()).To(Equal(2))
					Expect(fakeCake.RemoveArgsForCall(0)).To(Equal(layercake.DockerImageID("child1")))
					Expect(fakeCake.RemoveArgsForCall(1)).To(Equal(layercake.DockerImageID("child2")))
				})

			})

			Context("when getting the list of leaves fails", func() {
				It("returns the error", func() {
					fakeCake.GetAllLeavesReturns(nil, errors.New("firey potato"))
					Expect(gc.GC(logger, fakeCake)).To(MatchError("firey potato"))
				})
			})
		})
	})

	Context("when the threshold is not exceeded", func() {
		BeforeEach(func() {
			fakeCake.GetAllLeavesReturns([]layercake.ID{layercake.DockerImageID("child1"), layercake.DockerImageID("child2")}, nil)
			fakeThreshold.ExceededReturns(false)
		})

		It("it does not clean up anything", func() {
			Expect(gc.GC(logger, fakeCake)).To(Succeed())
			Expect(fakeCake.RemoveCallCount()).To(Equal(0))
		})
	})
})
