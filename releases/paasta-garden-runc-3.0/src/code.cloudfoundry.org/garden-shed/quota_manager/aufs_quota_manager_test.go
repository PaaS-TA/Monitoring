package quota_manager_test

import (
	"errors"

	"code.cloudfoundry.org/garden-shed/quota_manager"
	fakes "code.cloudfoundry.org/garden-shed/quota_manager/quota_managerfakes"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Quota Manager", func() {
	var (
		fakeBaseSizer *fakes.FakeBaseSizer
		fakeDiffSizer *fakes.FakeDiffSizer

		qm *quota_manager.AUFSQuotaManager
	)

	BeforeEach(func() {
		fakeBaseSizer = new(fakes.FakeBaseSizer)
		fakeDiffSizer = new(fakes.FakeDiffSizer)

		qm = &quota_manager.AUFSQuotaManager{
			BaseSizer: fakeBaseSizer,
			DiffSizer: fakeDiffSizer,
		}
	})

	Describe("GetUsage", func() {
		BeforeEach(func() {
			fakeBaseSizer.BaseSizeReturns(9876, nil)
			fakeDiffSizer.DiffSizeReturns(12345, nil)
		})

		It("returns the exclusive bytes used based on the Diff Size", func() {
			usage, err := qm.GetUsage(lagertest.NewTestLogger("test"), "some/path")
			Expect(err).NotTo(HaveOccurred())
			Expect(usage.ExclusiveBytesUsed).To(BeEquivalentTo(12345))
		})

		It("returns an error if finding the diff size fails", func() {
			fakeDiffSizer.DiffSizeReturns(12345, errors.New("something something"))
			_, err := qm.GetUsage(lagertest.NewTestLogger("test"), "some/path")
			Expect(err).To(MatchError("something something"))
		})

		It("returns the total bytes used by adding the Diff Size to the Base Size", func() {
			usage, err := qm.GetUsage(lagertest.NewTestLogger("test"), "some/path")
			Expect(err).NotTo(HaveOccurred())
			Expect(usage.TotalBytesUsed).To(BeEquivalentTo(12345 + 9876))
		})
	})
})
