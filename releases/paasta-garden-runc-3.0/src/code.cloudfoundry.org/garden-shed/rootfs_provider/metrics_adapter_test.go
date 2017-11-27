package rootfs_provider

import (
	"errors"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metrics Adapter", func() {
	It("converts metrics calls using ID on to GetUsage calls using path", func() {
		mAdapter := MetricsAdapter{
			fn: func(logger lager.Logger, rootfsPath string) (garden.ContainerDiskStat, error) {
				Expect(rootfsPath).To(Equal("/foo/bar/banana"))
				return garden.ContainerDiskStat{
					TotalBytesUsed: 12,
				}, errors.New("potato")
			},
			id2path: func(id layercake.ID) string {
				return "/foo/bar/" + id.GraphID()
			},
		}

		stat, err := mAdapter.Metrics(lagertest.NewTestLogger("test"), layercake.DockerImageID("banana"))
		Expect(err).To(MatchError("potato"))
		Expect(stat.TotalBytesUsed).To(BeEquivalentTo(12))
	})
})
